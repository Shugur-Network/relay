package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/Shugur-Network/relay/internal/logger"
	"github.com/Shugur-Network/relay/internal/relay/nips"
	"github.com/jackc/pgx/v5"
	nostr "github.com/nbd-wtf/go-nostr"
	"go.uber.org/zap"
)

// GetEvents retrieves events based on Nostr filters
func (db *DB) GetEvents(ctx context.Context, filter nostr.Filter) ([]nostr.Event, error) {
	// Compile the filter for efficient processing
	cf := CompileFilter(filter)

	// Build the optimized query
	query, args, err := cf.BuildQuery()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	// Create context with timeout
	queryCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Log the query for debugging
	logger.Debug("Executing query",
		zap.String("query", query),
		zap.Int("arg_count", len(args)))

	// Execute query
	rows, err := db.Pool.Query(queryCtx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	// Pre-allocate results slice with a reasonable upper bound to avoid excessive allocation
	const maxQueryPrealloc = 1000
	capHint := cf.Limit
	if capHint < 0 {
		capHint = 0
	}
	if capHint > maxQueryPrealloc {
		capHint = maxQueryPrealloc
	}
	events := make([]nostr.Event, 0, capHint)

	// Process rows
	for rows.Next() {
		var evt nostr.Event
		var createdAt int64
		var rawTags []byte

		if err := rows.Scan(&evt.ID, &evt.PubKey, &evt.Kind, &createdAt, &evt.Content, &rawTags, &evt.Sig); err != nil {
			logger.Warn("Row scan failed", zap.Error(err))
			continue
		}

		evt.CreatedAt = nostr.Timestamp(createdAt)

		// Parse tags
		if len(rawTags) > 0 {
			if err := json.Unmarshal(rawTags, &evt.Tags); err != nil {
				logger.Warn("Failed to unmarshal tags", zap.Error(err))
				evt.Tags = []nostr.Tag{}
			}
		}

		events = append(events, evt)
	}

	// Reorder events in ascending order by created_at
	sort.Slice(events, func(i, j int) bool {
		return events[i].CreatedAt < events[j].CreatedAt
	})

	return events, nil
}

// GetEventByID retrieves a single event by its ID.
func (db *DB) GetEventByID(ctx context.Context, eventID string) (nostr.Event, error) {
	query := `SELECT id, pubkey, kind, created_at, content, tags, sig FROM events WHERE id = $1`
	row := db.Pool.QueryRow(ctx, query, eventID)

	var evt nostr.Event
	var createdAt int64
	err := row.Scan(&evt.ID, &evt.PubKey, &evt.Kind, &createdAt, &evt.Content, &evt.Tags, &evt.Sig)
	if err != nil {
		return nostr.Event{}, fmt.Errorf("event not found: %w", err)
	}

	evt.CreatedAt = nostr.Timestamp(createdAt) // Convert Unix timestamp to nostr.Timestamp

	return evt, nil
}

// InsertEvent directly inserts a single event
func (db *DB) InsertEvent(ctx context.Context, evt nostr.Event) error {

	// Check Bloom filter first to avoid duplicate DB operations
	if db.Bloom.Test([]byte(evt.ID)) {
		// Already have this event

		return nil
	}
	// No need to add to Bloom filter here - that should be handled by the caller
	// so that we can control when the event is considered "processed"

	_, err := db.Pool.Exec(ctx,
		`INSERT INTO events (id, pubkey, created_at, kind, tags, content, sig)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 ON CONFLICT (id) DO NOTHING`,
		evt.ID, evt.PubKey, evt.CreatedAt.Time().Unix(),
		evt.Kind, evt.Tags, evt.Content, evt.Sig)

	if err != nil {
		return fmt.Errorf("failed to insert event: %w", err)
	}

	return nil
}

// Modified EventBuffer that processes events one at a time

// BatchInsertEvents optimized for CockroachDB with timeout handling
func (db *DB) BatchInsertEvents(ctx context.Context, events []nostr.Event) error {
	if len(events) == 0 {
		return nil
	}

	// Use smaller batches for CockroachDB
	const batchSize = 50

	// Execute in smaller batches with retries
	for i := 0; i < len(events); i += batchSize {
		end := i + batchSize
		if end > len(events) {
			end = len(events)
		}

		batchEvents := events[i:end]
		err := db.executeWithRetry(ctx, func(retryCtx context.Context) error {
			return db.insertEventBatch(retryCtx, batchEvents)
		})

		if err != nil {
			return fmt.Errorf("batch insert failed: %w", err)
		}
	}

	return nil
}

// Helper for actual batch insertion
func (db *DB) insertEventBatch(ctx context.Context, events []nostr.Event) error {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			db.recordError(fmt.Errorf("rollback failed: %w", rollbackErr))
		}
	}()

	// Set statement timeout and priority
	_, err = tx.Exec(ctx, "SET TRANSACTION PRIORITY HIGH")
	if err != nil {
		return fmt.Errorf("failed to set transaction priority: %w", err)
	}

	batch := &pgx.Batch{}
	for _, evt := range events {
		// Add event to bloom filter first
		db.Bloom.AddString(evt.ID)

		batch.Queue(
			`INSERT INTO events (id, pubkey, created_at, kind, tags, content, sig)
             VALUES ($1, $2, $3, $4, $5, $6, $7)
             ON CONFLICT (id) DO NOTHING`,
			evt.ID,
			evt.PubKey,
			evt.CreatedAt.Time().Unix(),
			evt.Kind,
			evt.Tags,
			evt.Content,
			evt.Sig,
		)
	}

	results := tx.SendBatch(ctx, batch)
	if err := results.Close(); err != nil {
		return fmt.Errorf("batch execution failed: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("transaction commit failed: %w", err)
	}

	return nil
}

// GetReplaceableEvent retrieves the latest replaceable event for a given pubkey and kind.
func (db *DB) GetReplaceableEvent(ctx context.Context, pubkey string, kind int) (nostr.Event, error) {
	query := `
		SELECT id, pubkey, kind, created_at, content, tags, sig
		FROM events
		WHERE pubkey = $1 AND kind = $2
		ORDER BY created_at DESC
		LIMIT 1`

	row, err := db.ExecuteQuery(ctx, query, pubkey, kind)
	if err != nil {
		return nostr.Event{}, fmt.Errorf("failed to fetch replaceable event: %w", err)
	}

	var evt nostr.Event
	var createdAt int64
	err = row.Scan(&evt.ID, &evt.PubKey, &evt.Kind, &createdAt, &evt.Content, &evt.Tags, &evt.Sig)
	if err != nil {
		return nostr.Event{}, fmt.Errorf("replaceable event not found: %w", err)
	}

	evt.CreatedAt = nostr.Timestamp(createdAt) // Convert Unix timestamp to nostr.Timestamp

	return evt, nil
}

// GetAddressableEvent retrieves the latest addressable event for a given pubkey, kind, and 'd' tag.
func (db *DB) GetAddressableEvent(ctx context.Context, pubkey string, kind int, dVal string) (nostr.Event, error) {
	query := `
		SELECT id, pubkey, kind, created_at, content, tags, sig
		FROM events
		WHERE pubkey = $1 AND kind = $2 AND tags @> $3
		ORDER BY created_at DESC
		LIMIT 1`

	row, err := db.ExecuteQuery(ctx, query, pubkey, kind, fmt.Sprintf(`[["d", "%s"]]`, dVal))
	if err != nil {
		return nostr.Event{}, fmt.Errorf("failed to fetch addressable event: %w", err)
	}

	var evt nostr.Event
	var createdAt int64
	err = row.Scan(&evt.ID, &evt.PubKey, &evt.Kind, &createdAt, &evt.Content, &evt.Tags, &evt.Sig)
	if err != nil {
		return nostr.Event{}, fmt.Errorf("addressable event not found: %w", err)
	}

	evt.CreatedAt = nostr.Timestamp(createdAt) // Convert Unix timestamp to nostr.Timestamp

	return evt, nil
}

// DeleteExpiredEvents removes events that have expired based on the "expiration" tag.
func (db *DB) DeleteExpiredEvents(ctx context.Context) error {
	query := `
		DELETE FROM events
		WHERE EXISTS (
			SELECT 1 FROM jsonb_array_elements(tags) AS tag
			WHERE tag->>0 = 'expiration' 
			AND tag->>1 IS NOT NULL 
			AND (tag->>1)::BIGINT < extract(epoch FROM now())
		)`

	logger.Debug("🗑 Deleting expired events...")

	_, err := db.Pool.Exec(ctx, query)
	if err != nil {
		logger.Error("❌ Failed to delete expired events", zap.Error(err))
		return fmt.Errorf("failed to delete expired events: %w", err)
	}

	logger.Debug("✅ Expired events deleted successfully")
	return nil
}

// CleanExpiredEvents removes events with expiration tags that have passed their expiration time
func (db *DB) CleanExpiredEvents(ctx context.Context) (int, error) {
	if !db.isConnected() {
		return 0, fmt.Errorf("database is not connected")
	}

	logger.Debug("Deleting expired events...")

	query := `
		DELETE FROM events
		WHERE EXISTS (
			SELECT 1 FROM jsonb_array_elements(tags) AS tag
			WHERE tag->>0 = 'expiration' 
			AND tag->>1 IS NOT NULL 
			AND (tag->>1)::BIGINT <= $1
		)
	`

	result, err := db.Pool.Exec(ctx, query, time.Now().Unix())
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired events: %w", err)
	}

	count := result.RowsAffected()
	logger.Debug("Expired events deleted",
		zap.Int64("count", count))

	return int(count), nil
}

// StartExpiredEventsCleaner starts a background goroutine to clean expired events periodically
func (db *DB) StartExpiredEventsCleaner(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				logger.Debug("Running expired events cleanup...")
				count, err := db.CleanExpiredEvents(ctx)
				if err != nil {
					logger.Error("Failed to clean expired events", zap.Error(err))
				} else if count > 0 {
					logger.Info("Cleaned expired events", zap.Int("count", count))
				}
			}
		}
	}()
}

// GetEventCount returns the count of events matching the given filter
func (db *DB) GetEventCount(ctx context.Context, filter nostr.Filter) (int64, error) {
	// PERFORMANCE: Create a query builder with reasonable capacity
	query := strings.Builder{}
	query.Grow(256) // Pre-allocate string builder capacity
	args := make([]interface{}, 0, 10)
	argIndex := 1

	// Start with base SELECT COUNT
	query.WriteString(`SELECT COUNT(*) FROM events`)

	// Track if we need to add WHERE
	needsWhere := false
	addWhere := func() {
		if !needsWhere {
			query.WriteString(` WHERE `)
			needsWhere = true
		} else {
			query.WriteString(` AND `)
		}
	}

	// Add filters in order of index selectivity
	hasIDFilter := len(filter.IDs) > 0
	hasAuthorFilter := len(filter.Authors) > 0
	hasKindFilter := len(filter.Kinds) > 0
	hasSinceFilter := filter.Since != nil
	hasUntilFilter := filter.Until != nil

	// Apply filters based on most efficient index usage
	if hasIDFilter {
		// IDs are primary keys - most selective
		addWhere()
		query.WriteString(fmt.Sprintf("id = ANY($%d)", argIndex))
		args = append(args, filter.IDs)
		argIndex++
	}

	if hasAuthorFilter {
		addWhere()
		query.WriteString(fmt.Sprintf("pubkey = ANY($%d)", argIndex))
		args = append(args, filter.Authors)
		argIndex++
	}

	if hasKindFilter {
		addWhere()
		query.WriteString(fmt.Sprintf("kind = ANY($%d)", argIndex))
		args = append(args, filter.Kinds)
		argIndex++
	}

	// Always apply time filters after key/author filters
	if hasSinceFilter {
		addWhere()
		query.WriteString(fmt.Sprintf("created_at >= $%d", argIndex))
		args = append(args, filter.Since.Time().Unix())
		argIndex++
	}

	if hasUntilFilter {
		addWhere()
		query.WriteString(fmt.Sprintf("created_at <= $%d", argIndex))
		args = append(args, filter.Until.Time().Unix())
		argIndex++
	}

	// Handle tag filtering
	if len(filter.Tags) > 0 {
		for tagName, tagValues := range filter.Tags {
			if len(tagValues) > 0 {
				addWhere()
				// Use the inverted index on tags
				query.WriteString(fmt.Sprintf("tags @> $%d", argIndex))
				tagArray := make([][]string, len(tagValues))
				for i, val := range tagValues {
					tagArray[i] = []string{tagName, val}
				}
				args = append(args, tagArray)
				argIndex++
			}
		}
	}

	// Log the query for debugging
	logger.Debug("Executing count query",
		zap.String("query", query.String()),
		zap.Int("arg_count", len(args)))

	// Execute query with timeout
	var count int64
	err := db.Pool.QueryRow(ctx, query.String(), args...).Scan(&count)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return 0, fmt.Errorf("count operation timed out")
		}
		return 0, fmt.Errorf("failed to execute count query: %w", err)
	}

	return count, nil
}

func (db *DB) EventExists(ctx context.Context, eventID string) (bool, error) {
	var exists bool
	err := db.Pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM events WHERE id = $1)`,
		eventID,
	).Scan(&exists)
	return exists, err
}

func (db *DB) InsertReplaceableEvent(ctx context.Context, evt nostr.Event) error {
	// First, delete any existing replaceable event for this pubkey and kind
	_, err := db.Pool.Exec(ctx,
		`DELETE FROM events 
		 WHERE pubkey = $1 AND kind = $2`,
		evt.PubKey, evt.Kind)
	if err != nil {
		return fmt.Errorf("failed to delete old replaceable event: %w", err)
	}

	// Then insert the new event
	_, err = db.Pool.Exec(ctx,
		`INSERT INTO events (id, pubkey, created_at, kind, tags, content, sig)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		evt.ID, evt.PubKey, evt.CreatedAt.Time().Unix(),
		evt.Kind, evt.Tags, evt.Content, evt.Sig)
	if err != nil {
		return fmt.Errorf("failed to insert new replaceable event: %w", err)
	}

	// Add to Bloom filter
	db.Bloom.AddString(evt.ID)

	return nil
}

// InsertAddressableEvent upserts (pubkey, kind, dTag) = unique
func (db *DB) InsertAddressableEvent(ctx context.Context, evt nostr.Event) error {
	dVal := nips.GetTagValue(evt, "d")
	if dVal == "" {
		return db.InsertEvent(ctx, evt) // fallback
	}

	_, err := db.Pool.Exec(ctx,
		`DELETE FROM events 
         WHERE pubkey=$1 AND kind=$2 AND tags @> $3`,
		evt.PubKey, evt.Kind, fmt.Sprintf(`[["d","%s"]]`, dVal),
	)
	if err != nil {
		return err
	}

	_, err = db.Pool.Exec(ctx,
		`INSERT INTO events (id,pubkey,created_at,kind,tags,content,sig)
         VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		evt.ID, evt.PubKey, evt.CreatedAt.Time().Unix(),
		evt.Kind, evt.Tags, evt.Content, evt.Sig,
	)
	if err == nil {
		db.Bloom.AddString(evt.ID)
	}
	return err
}

func (db *DB) persistDeletion(ctx context.Context, del nostr.Event) error {
	var ids []string
	for _, t := range del.Tags {
		if len(t) >= 2 && t[0] == "e" {
			ids = append(ids, t[1])
		}
	}
	if len(ids) == 0 {
		return errors.New("deletion event without e‑tags")
	}

	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			db.recordError(fmt.Errorf("rollback failed: %w", rollbackErr))
		}
	}()

	// 1) delete only events OWNED by the deleter
	_, err = tx.Exec(ctx,
		`DELETE FROM events WHERE id = ANY($1) AND pubkey = $2`,
		ids, del.PubKey)
	if err != nil {
		return err
	}

	// 2) insert the deletion event itself
	_, err = tx.Exec(ctx,
		`INSERT INTO events (id,pubkey,created_at,kind,tags,content,sig)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		del.ID, del.PubKey, del.CreatedAt.Time().Unix(),
		del.Kind, del.Tags, del.Content, del.Sig)
	if err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	db.Bloom.AddString(del.ID)
	return nil
}

// GetTotalEventCount returns the total number of events stored in the database
func (db *DB) GetTotalEventCount(ctx context.Context) (int64, error) {
	if !db.isConnected() {
		return 0, fmt.Errorf("database is not connected")
	}

	var count int64
	err := db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM events").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get total event count: %w", err)
	}

	return count, nil
}

package relay

import (
	"context"
	"time"

	"github.com/Shugur-Network/relay/internal/logger"
	"github.com/Shugur-Network/relay/internal/metrics"
	"github.com/Shugur-Network/relay/internal/relay/nips"
	nostr "github.com/nbd-wtf/go-nostr"
	"go.uber.org/zap"
)

func (c *WsConnection) handleRequest(ctx context.Context, arr []interface{}) {
	// Log the start of request processing
	logger.Debug("Processing REQ command",
		zap.String("client", c.ws.RemoteAddr().String()))

	// Validate array length
	if len(arr) < 3 {
		logger.Warn("Invalid REQ command: missing subscription ID or filter",
			zap.String("client", c.ws.RemoteAddr().String()))
		c.sendNotice("REQ command missing subscription ID or filter")
		return
	}

	// Extract subscription ID
	subID, ok := arr[1].(string)
	if !ok || subID == "" {
		logger.Warn("Invalid REQ command: subscription ID must be a string",
			zap.String("client", c.ws.RemoteAddr().String()))
		c.sendNotice("REQ command subscription ID must be a string")
		return
	}

	// Validate subscription ID length
	if len(subID) > 64 {
		c.sendNotice("Subscription ID too long (max 64 chars)")
		return
	}

	// Remove existing subscription if present
	if c.hasSubscription(subID) {
		logger.Debug("Replacing existing subscription",
			zap.String("sub_id", subID),
			zap.String("client", c.ws.RemoteAddr().String()))
		c.removeSubscription(subID)
	}

	// Parse the filter with support for #tag syntax
	var f nostr.Filter
	if len(arr) >= 3 {
		filter, err := parseFilterFromRaw(arr[2])
		if err != nil {
			logger.Warn("Failed to parse filter",
				zap.String("sub_id", subID),
				zap.Error(err),
				zap.String("client", c.ws.RemoteAddr().String()))
			c.sendNotice("Invalid filter: " + err.Error())
			return
		}
		f = filter
	} else {
		c.sendNotice("REQ command missing filter")
		return
	}

	// Apply cap to limit if needed
	if f.Limit <= 0 || f.Limit > 500 {
		f.Limit = 500
	}

	// Validate filter with the validator
	if err := c.node.GetValidator().ValidateFilter(f); err != nil {
		logger.Warn("Filter validation failed",
			zap.String("sub_id", subID),
			zap.Error(err),
			zap.String("client", c.ws.RemoteAddr().String()))
		c.sendClosed(subID, nips.FormatErrorMessage(nips.ErrorCodeInvalidFilter, err.Error()))
		return
	}

	// Check special validation for specific filter types
	if len(f.Kinds) > 0 {
		switch {
		case containsKind(f.Kinds, nips.KindRelayList):
			if err := nips.ValidateRelayListFilter(f); err != nil {
				c.sendClosed(subID, nips.FormatErrorMessage(nips.ErrorCodeInvalidFilter, err.Error()))
				return
			}
		}
	}

	// Validate search if present
	if f.Search != "" {
		if err := nips.ValidateSearchFilter(f, nips.DefaultSearchOptions()); err != nil {
			c.sendClosed(subID, nips.FormatErrorMessage(nips.ErrorCodeInvalidFilter, err.Error()))
			return
		}
	}

	// Store subscription
	c.addSubscription(subID, []nostr.Filter{f})

	// Update metrics
	metrics.ActiveSubscriptions.Inc()

	// Query DB and send events in a goroutine
	go c.processSubscription(ctx, subID, f)
}

// processSubscription handles the database query and sending events to the client
func (c *WsConnection) processSubscription(ctx context.Context, subID string, f nostr.Filter) {
	// Create a context with timeout for the query
	_, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Query events from the database
	start := time.Now()
	events, err := c.QueryEvents(ctx, f)
	duration := time.Since(start)

	// Log query performance
	logger.Debug("Query execution completed",
		zap.String("sub_id", subID),
		zap.Duration("duration", duration),
		zap.Int("events_count", len(events)),
		zap.String("client", c.ws.RemoteAddr().String()))

	if err != nil {
		logger.Error("Failed to query events",
			zap.String("sub_id", subID),
			zap.Error(err),
			zap.String("client", c.ws.RemoteAddr().String()))
		c.sendNotice(nips.ErrDatabaseError)
		return
	}

	// Check if client is still connected before proceeding
	if c.isClosed.Load() {
		return
	}

	// Apply special validation for specific event kinds
	if len(f.Kinds) == 1 {
		switch f.Kinds[0] {
		case nips.KindRelayList:
			// Filter out invalid relay list events
			validEvents := make([]nostr.Event, 0, len(events))
			for _, evt := range events {
				if err := nips.ValidateRelayListEvent(evt); err == nil {
					validEvents = append(validEvents, evt)
				}
			}
			events = validEvents
		}
	}

	// Send events to the client
	sentCount := 0
	for _, evt := range events {
		// Check again if client is still connected
		if c.isClosed.Load() {
			return
		}

		// For DMs, check if client is authorized
		// Note: Gift wrap events (1059) are excluded as they handle access control via encryption
		if evt.Kind == 4 || evt.Kind == 14 || evt.Kind == 15 {
			if !isAuthorizedForDM(&evt, c.getSubscriptionFilters(subID)) {
				continue // Skip sending this event
			}
		}

		// Send the event
		c.SendEvent(subID, &evt)
		sentCount++
	}

	logger.Debug("Subscription events sent",
		zap.String("sub_id", subID),
		zap.Int("sent_count", sentCount),
		zap.String("client", c.ws.RemoteAddr().String()))

	// Send EOSE (End of Stored Events)
	if !c.isClosed.Load() {
		c.sendEOSE(subID)
	}
}

// isAuthorizedForDM checks if a client should receive a DM
func isAuthorizedForDM(evt *nostr.Event, filters []nostr.Filter) bool {
	// Skip authorization for non-DM events
	// Note: Gift wrap events (1059) are excluded as they handle access control via encryption
	if evt.Kind != 4 && evt.Kind != 14 && evt.Kind != 15 {
		return true
	}

	// Extract DM recipients from event tags
	recipients := make(map[string]bool)
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == "p" {
			recipients[tag[1]] = true
		}
	}
	// Sender can see their own DMs
	recipients[evt.PubKey] = true

	// Check if any filter authorizes the client to see this DM
	for _, filter := range filters {
		// If filter explicitly includes the event's author
		for _, author := range filter.Authors {
			if author == evt.PubKey {
				return true
			}
		}

		// If filter explicitly includes the client's pubkey as a recipient
		if pTags, ok := filter.Tags["p"]; ok {
			for _, pubkey := range pTags {
				if recipients[pubkey] {
					return true
				}
			}
		}
	}
	return false
}

func (c *WsConnection) handleClose(arr []interface{}) {
	// Log the start of close processing
	logger.Debug("Processing CLOSE command",
		zap.String("client", c.ws.RemoteAddr().String()))

	// Validate array length
	if len(arr) < 2 {
		logger.Warn("Invalid CLOSE command: missing subscription ID",
			zap.String("client", c.ws.RemoteAddr().String()))
		c.sendNotice("CLOSE command missing subscription ID")
		return
	}

	// Extract and validate subscription ID
	subID, ok := arr[1].(string)
	if !ok {
		logger.Warn("Invalid CLOSE command: subscription ID must be a string",
			zap.String("client", c.ws.RemoteAddr().String()))
		c.sendNotice("CLOSE command subscription ID must be a string")
		return
	}

	// Check if subscription exists before attempting to close
	if !c.hasSubscription(subID) {
		logger.Debug("Attempted to close non-existent subscription",
			zap.String("sub_id", subID),
			zap.String("client", c.ws.RemoteAddr().String()))
		c.sendClosed(subID, "subscription not found")
		return
	}

	// Log subscription closure
	logger.Debug("Closing subscription",
		zap.String("sub_id", subID),
		zap.String("client", c.ws.RemoteAddr().String()))

	// Remove subscription and send confirmation
	c.removeSubscription(subID)
	c.sendClosed(subID, "subscription closed")

	// Update metrics
	metrics.ActiveSubscriptions.Dec()

	// Log successful closure
	logger.Debug("Subscription successfully closed",
		zap.String("sub_id", subID),
		zap.String("client", c.ws.RemoteAddr().String()))
}

// handleCountRequest processes COUNT commands for NIP-45
func (c *WsConnection) handleCountRequest(ctx context.Context, arr []interface{}) {
	// Log the start of count request processing
	logger.Debug("Starting count request processing",
		zap.String("client", c.ws.RemoteAddr().String()))

	// Parse the COUNT command using NIP-45 module
	countCmd, err := nips.ParseCountCommand(arr)
	if err != nil {
		logger.Warn("Invalid COUNT command",
			zap.Error(err),
			zap.String("client", c.ws.RemoteAddr().String()))
		c.sendNotice("Invalid COUNT command: " + err.Error())
		return
	}

	// Parse the filter using existing parseFilterFromRaw
	if len(arr) >= 3 {
		filter, err := parseFilterFromRaw(arr[2])
		if err != nil {
			logger.Warn("Failed to parse filter for COUNT",
				zap.String("sub_id", countCmd.SubID),
				zap.Error(err),
				zap.String("client", c.ws.RemoteAddr().String()))
			c.sendNotice("Invalid filter: " + err.Error())
			return
		}
		countCmd.Filter = filter
	} else {
		c.sendNotice("COUNT command missing filter")
		return
	}

	// Process count in a goroutine
	go func() {
		// Create a context with timeout for the count operation
		countCtx, cancel := context.WithTimeout(ctx, nips.CountTimeout)
		defer cancel()

		// Validate the filter using NIP-45
		_, err := nips.HandleCountRequest(countCtx, countCmd.SubID, countCmd.Filter)
		if err != nil {
			logger.Warn("COUNT filter validation failed",
				zap.String("sub_id", countCmd.SubID),
				zap.Error(err),
				zap.String("client", c.ws.RemoteAddr().String()))
			c.sendNotice("Invalid COUNT filter: " + err.Error())
			return
		}

		// Get count from database
		start := time.Now()
		count, err := c.node.DB().GetEventCount(countCtx, countCmd.Filter)
		duration := time.Since(start)

		// Check if client is still connected
		if c.isClosed.Load() {
			return
		}

		// Handle error
		if err != nil {
			logger.Error("COUNT request failed",
				zap.String("sub_id", countCmd.SubID),
				zap.Error(err),
				zap.String("client", c.ws.RemoteAddr().String()))
			c.sendNotice("error: count operation failed")
			return
		}

		// Log performance
		logger.Debug("Count operation completed",
			zap.String("sub_id", countCmd.SubID),
			zap.Duration("duration", duration),
			zap.Int64("count", count),
			zap.String("client", c.ws.RemoteAddr().String()))

		// Send the count response (NIP-45 format)
		response := &nips.CountResponse{Count: count}
		c.sendMessage("COUNT", countCmd.SubID, response)
	}()
}

// Subscription management helpers
func (c *WsConnection) hasSubscription(subID string) bool {
	c.subMu.RLock()
	defer c.subMu.RUnlock()
	_, ok := c.subscriptions[subID]
	return ok
}

func (c *WsConnection) addSubscription(subID string, filters []nostr.Filter) {
	c.subMu.Lock()
	defer c.subMu.Unlock()
	c.subscriptions[subID] = filters
}

func (c *WsConnection) removeSubscription(subID string) {
	c.subMu.Lock()
	defer c.subMu.Unlock()
	delete(c.subscriptions, subID)
}

func (c *WsConnection) getSubscriptionFilters(subID string) []nostr.Filter {
	c.subMu.RLock()
	defer c.subMu.RUnlock()
	filters, ok := c.subscriptions[subID]
	if !ok {
		return nil
	}
	return filters
}

// GetSubscriptions returns all active subscriptions
func (c *WsConnection) GetSubscriptions() map[string][]nostr.Filter {
	c.subMu.RLock()
	defer c.subMu.RUnlock()

	// Create a copy to avoid concurrent access issues
	cp := make(map[string][]nostr.Filter, len(c.subscriptions))
	for k, v := range c.subscriptions {
		cp[k] = v
	}
	return cp
}

// SendEvent sends a Nostr event to the client for a specific subscription
func (c *WsConnection) SendEvent(subID string, evt *nostr.Event) {
	// Check if subscription exists
	if !c.HasSubscription(subID) {
		return
	}

	// Send the event
	c.sendMessage("EVENT", subID, evt)
}

// containsKind checks if a slice of kinds contains a specific kind
func containsKind(kinds []int, kind int) bool {
	for _, k := range kinds {
		if k == kind {
			return true
		}
	}
	return false
}

package nips

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	nostr "github.com/nbd-wtf/go-nostr"
	"github.com/Shugur-Network/relay/internal/logger"
	"go.uber.org/zap"
)

// ValidateDateBasedCalendarEvent validates NIP-52 date-based calendar events (kind 31922)
func ValidateDateBasedCalendarEvent(event *nostr.Event) error {
	// Basic validation
	if event == nil {
		return fmt.Errorf("event is nil")
	}

	logger.Debug("NIP-52: Validating date-based calendar event",
		zap.String("event_id", event.ID),
		zap.String("pubkey", event.PubKey))

	if event.Kind != 31922 {
		return fmt.Errorf("invalid kind for date-based calendar event: expected 31922, got %d", event.Kind)
	}

	// Validate required and optional tags
	if err := validateDateBasedCalendarEventTags(event); err != nil {
		return fmt.Errorf("invalid date-based calendar event tags: %w", err)
	}

	logger.Debug("NIP-52: Date-based calendar event validation successful",
		zap.String("event_id", event.ID))
	return nil
}

// ValidateTimeBasedCalendarEvent validates NIP-52 time-based calendar events (kind 31923)
func ValidateTimeBasedCalendarEvent(event *nostr.Event) error {
	// Basic validation
	if event == nil {
		return fmt.Errorf("event is nil")
	}

	logger.Debug("NIP-52: Validating time-based calendar event",
		zap.String("event_id", event.ID),
		zap.String("pubkey", event.PubKey))

	if event.Kind != 31923 {
		return fmt.Errorf("invalid kind for time-based calendar event: expected 31923, got %d", event.Kind)
	}

	// Validate required and optional tags
	if err := validateTimeBasedCalendarEventTags(event); err != nil {
		return fmt.Errorf("invalid time-based calendar event tags: %w", err)
	}

	logger.Debug("NIP-52: Time-based calendar event validation successful",
		zap.String("event_id", event.ID))
	return nil
}

// ValidateCalendar validates NIP-52 calendar events (kind 31924)
func ValidateCalendar(event *nostr.Event) error {
	// Basic validation
	if event == nil {
		return fmt.Errorf("event is nil")
	}

	logger.Debug("NIP-52: Validating calendar event",
		zap.String("event_id", event.ID),
		zap.String("pubkey", event.PubKey))

	if event.Kind != 31924 {
		return fmt.Errorf("invalid kind for calendar: expected 31924, got %d", event.Kind)
	}

	// Validate required and optional tags
	if err := validateCalendarTags(event); err != nil {
		return fmt.Errorf("invalid calendar tags: %w", err)
	}

	logger.Debug("NIP-52: Calendar validation successful",
		zap.String("event_id", event.ID))
	return nil
}

// ValidateCalendarEventRSVP validates NIP-52 calendar event RSVP events (kind 31925)
func ValidateCalendarEventRSVP(event *nostr.Event) error {
	// Basic validation
	if event == nil {
		return fmt.Errorf("event is nil")
	}

	logger.Debug("NIP-52: Validating calendar event RSVP",
		zap.String("event_id", event.ID),
		zap.String("pubkey", event.PubKey))

	if event.Kind != 31925 {
		return fmt.Errorf("invalid kind for calendar event RSVP: expected 31925, got %d", event.Kind)
	}

	// Validate required and optional tags
	if err := validateCalendarEventRSVPTags(event); err != nil {
		return fmt.Errorf("invalid calendar event RSVP tags: %w", err)
	}

	logger.Debug("NIP-52: Calendar event RSVP validation successful",
		zap.String("event_id", event.ID))
	return nil
}

// validateDateBasedCalendarEventTags validates tags for date-based calendar events
func validateDateBasedCalendarEventTags(event *nostr.Event) error {
	var hasDTag bool
	var hasTitleTag bool
	var hasStartTag bool
	var startDate string
	var endDate string

	for _, tag := range event.Tags {
		if len(tag) == 0 {
			continue
		}

		switch tag[0] {
		case "d":
			if err := validateCalendarDTag(tag); err != nil {
				return err
			}
			hasDTag = true
		case "title":
			if err := validateCalendarTitleTag(tag); err != nil {
				return err
			}
			hasTitleTag = true
		case "start":
			if err := validateDateTag(tag, "start"); err != nil {
				return err
			}
			hasStartTag = true
			if len(tag) >= 2 {
				startDate = tag[1]
			}
		case "end":
			if err := validateDateTag(tag, "end"); err != nil {
				return err
			}
			if len(tag) >= 2 {
				endDate = tag[1]
			}
		case "summary":
			if err := validateCalendarSummaryTag(tag); err != nil {
				return err
			}
		case "image":
			if err := validateCalendarImageTag(tag); err != nil {
				return err
			}
		case "location":
			if err := validateCalendarLocationTag(tag); err != nil {
				return err
			}
		case "g":
			if err := validateGeohashTag(tag); err != nil {
				return err
			}
		case "p":
			if err := validateCalendarParticipantTag(tag); err != nil {
				return err
			}
		case "t":
			if err := validateCalendarHashtagTag(tag); err != nil {
				return err
			}
		case "r":
			if err := validateCalendarReferenceTag(tag); err != nil {
				return err
			}
		case "a":
			if err := validateCalendarATag(tag); err != nil {
				return err
			}
		case "name":
			// Deprecated, but still allowed
			if err := validateCalendarNameTag(tag); err != nil {
				return err
			}
		default:
			// Other tags are allowed
		}
	}

	// Required tags validation
	if !hasDTag {
		return fmt.Errorf("date-based calendar event must have a d tag")
	}

	if !hasTitleTag {
		return fmt.Errorf("date-based calendar event must have a title tag")
	}

	if !hasStartTag {
		return fmt.Errorf("date-based calendar event must have a start tag")
	}

	// Validate date ordering if both start and end are present
	if startDate != "" && endDate != "" {
		if err := validateDateOrdering(startDate, endDate); err != nil {
			return err
		}
	}

	return nil
}

// validateTimeBasedCalendarEventTags validates tags for time-based calendar events
func validateTimeBasedCalendarEventTags(event *nostr.Event) error {
	var hasDTag bool
	var hasTitleTag bool
	var hasStartTag bool
	var startTime int64
	var endTime int64

	for _, tag := range event.Tags {
		if len(tag) == 0 {
			continue
		}

		switch tag[0] {
		case "d":
			if err := validateCalendarDTag(tag); err != nil {
				return err
			}
			hasDTag = true
		case "title":
			if err := validateCalendarTitleTag(tag); err != nil {
				return err
			}
			hasTitleTag = true
		case "start":
			if err := validateTimestampTag(tag, "start"); err != nil {
				return err
			}
			hasStartTag = true
			if len(tag) >= 2 {
				if timestamp, err := strconv.ParseInt(tag[1], 10, 64); err == nil {
					startTime = timestamp
				}
			}
		case "end":
			if err := validateTimestampTag(tag, "end"); err != nil {
				return err
			}
			if len(tag) >= 2 {
				if timestamp, err := strconv.ParseInt(tag[1], 10, 64); err == nil {
					endTime = timestamp
				}
			}
		case "start_tzid":
			if err := validateTimezoneTag(tag, "start_tzid"); err != nil {
				return err
			}
		case "end_tzid":
			if err := validateTimezoneTag(tag, "end_tzid"); err != nil {
				return err
			}
		case "summary":
			if err := validateCalendarSummaryTag(tag); err != nil {
				return err
			}
		case "image":
			if err := validateCalendarImageTag(tag); err != nil {
				return err
			}
		case "location":
			if err := validateCalendarLocationTag(tag); err != nil {
				return err
			}
		case "g":
			if err := validateGeohashTag(tag); err != nil {
				return err
			}
		case "p":
			if err := validateCalendarParticipantTag(tag); err != nil {
				return err
			}
		case "t":
			if err := validateCalendarHashtagTag(tag); err != nil {
				return err
			}
		case "r":
			if err := validateCalendarReferenceTag(tag); err != nil {
				return err
			}
		case "a":
			if err := validateCalendarATag(tag); err != nil {
				return err
			}
		case "name":
			// Deprecated, but still allowed
			if err := validateCalendarNameTag(tag); err != nil {
				return err
			}
		default:
			// Other tags are allowed
		}
	}

	// Required tags validation
	if !hasDTag {
		return fmt.Errorf("time-based calendar event must have a d tag")
	}

	if !hasTitleTag {
		return fmt.Errorf("time-based calendar event must have a title tag")
	}

	if !hasStartTag {
		return fmt.Errorf("time-based calendar event must have a start tag")
	}

	// Validate timestamp ordering if both start and end are present
	if startTime > 0 && endTime > 0 {
		if startTime >= endTime {
			return fmt.Errorf("start timestamp must be less than end timestamp")
		}
	}

	return nil
}

// validateCalendarTags validates tags for calendar events
func validateCalendarTags(event *nostr.Event) error {
	var hasDTag bool
	var hasTitleTag bool

	for _, tag := range event.Tags {
		if len(tag) == 0 {
			continue
		}

		switch tag[0] {
		case "d":
			if err := validateCalendarDTag(tag); err != nil {
				return err
			}
			hasDTag = true
		case "title":
			if err := validateCalendarTitleTag(tag); err != nil {
				return err
			}
			hasTitleTag = true
		case "a":
			if err := validateCalendarEventReferenceTag(tag); err != nil {
				return err
			}
		default:
			// Other tags are allowed
		}
	}

	// Required tags validation
	if !hasDTag {
		return fmt.Errorf("calendar must have a d tag")
	}

	if !hasTitleTag {
		return fmt.Errorf("calendar must have a title tag")
	}

	return nil
}

// validateCalendarEventRSVPTags validates tags for calendar event RSVP events
func validateCalendarEventRSVPTags(event *nostr.Event) error {
	var hasDTag bool
	var hasATag bool
	var hasStatusTag bool

	for _, tag := range event.Tags {
		if len(tag) == 0 {
			continue
		}

		switch tag[0] {
		case "d":
			if err := validateCalendarDTag(tag); err != nil {
				return err
			}
			hasDTag = true
		case "a":
			if err := validateRSVPATag(tag); err != nil {
				return err
			}
			hasATag = true
		case "e":
			if err := validateRSVPETag(tag); err != nil {
				return err
			}
		case "status":
			if err := validateRSVPStatusTag(tag); err != nil {
				return err
			}
			hasStatusTag = true
		case "fb":
			if err := validateRSVPFreebusyTag(tag); err != nil {
				return err
			}
		case "p":
			if err := validateRSVPPTag(tag); err != nil {
				return err
			}
		default:
			// Other tags are allowed
		}
	}

	// Required tags validation
	if !hasDTag {
		return fmt.Errorf("calendar event RSVP must have a d tag")
	}

	if !hasATag {
		return fmt.Errorf("calendar event RSVP must have an a tag")
	}

	if !hasStatusTag {
		return fmt.Errorf("calendar event RSVP must have a status tag")
	}

	return nil
}

// validateCalendarDTag validates the d tag for calendar events
func validateCalendarDTag(tag nostr.Tag) error {
	if len(tag) != 2 {
		return fmt.Errorf("d tag must have exactly 2 elements")
	}

	identifier := tag[1]
	if identifier == "" {
		return fmt.Errorf("d tag identifier cannot be empty")
	}

	if len(identifier) > 200 {
		return fmt.Errorf("d tag identifier too long (max 200 characters)")
	}

	return nil
}

// validateCalendarTitleTag validates the title tag for calendar events
func validateCalendarTitleTag(tag nostr.Tag) error {
	if len(tag) != 2 {
		return fmt.Errorf("title tag must have exactly 2 elements")
	}

	title := tag[1]
	if title == "" {
		return fmt.Errorf("title cannot be empty")
	}

	if len(title) > 500 {
		return fmt.Errorf("title too long (max 500 characters)")
	}

	return nil
}

// validateCalendarSummaryTag validates the summary tag for calendar events
func validateCalendarSummaryTag(tag nostr.Tag) error {
	if len(tag) != 2 {
		return fmt.Errorf("summary tag must have exactly 2 elements")
	}

	summary := tag[1]
	if len(summary) > 1000 {
		return fmt.Errorf("summary too long (max 1000 characters)")
	}

	return nil
}

// validateCalendarImageTag validates the image tag for calendar events
func validateCalendarImageTag(tag nostr.Tag) error {
	if len(tag) != 2 {
		return fmt.Errorf("image tag must have exactly 2 elements")
	}

	imageURL := tag[1]
	if err := validateImageURL(imageURL); err != nil {
		return fmt.Errorf("invalid image URL: %w", err)
	}

	return nil
}

// validateCalendarLocationTag validates the location tag for calendar events
func validateCalendarLocationTag(tag nostr.Tag) error {
	if len(tag) != 2 {
		return fmt.Errorf("location tag must have exactly 2 elements")
	}

	location := tag[1]
	if len(location) > 500 {
		return fmt.Errorf("location too long (max 500 characters)")
	}

	return nil
}

// validateGeohashTag validates the g (geohash) tag for calendar events
func validateGeohashTag(tag nostr.Tag) error {
	if len(tag) != 2 {
		return fmt.Errorf("g tag must have exactly 2 elements")
	}

	geohash := tag[1]
	if geohash == "" {
		return fmt.Errorf("geohash cannot be empty")
	}

	// Geohash validation: should be base32 characters
	validGeohash := regexp.MustCompile(`^[0123456789bcdefghjkmnpqrstuvwxyz]+$`)
	if !validGeohash.MatchString(geohash) {
		return fmt.Errorf("invalid geohash format")
	}

	if len(geohash) > 12 {
		return fmt.Errorf("geohash too long (max 12 characters)")
	}

	return nil
}

// validateCalendarParticipantTag validates the p (participant) tag for calendar events
func validateCalendarParticipantTag(tag nostr.Tag) error {
	if len(tag) < 2 || len(tag) > 4 {
		return fmt.Errorf("p tag must have 2-4 elements")
	}

	// The p tag might have comma-separated values in one field, or separate fields
	pubkeyField := tag[1]
	parts := strings.Split(pubkeyField, ",")
	pubkey := parts[0]

	// Validate pubkey
	if len(pubkey) != 64 {
		return fmt.Errorf("participant pubkey must be 64 hex characters, got %d", len(pubkey))
	}

	if !isHexChar64(pubkey) {
		return fmt.Errorf("participant pubkey must be valid hex")
	}

	// If relay and role are in the same field (comma-separated)
	if len(parts) > 1 {
		// parts[1] is relay hint (optional)
		if len(parts) > 1 && parts[1] != "" {
			if err := validateRelayURL(parts[1]); err != nil {
				return fmt.Errorf("invalid participant relay hint: %w", err)
			}
		}
		// parts[2] is role (optional)
		if len(parts) > 2 && parts[2] != "" {
			role := parts[2]
			if len(role) > 100 {
				return fmt.Errorf("participant role too long (max 100 characters)")
			}
		}
	} else {
		// Optional relay hint validation (separate field)
		if len(tag) >= 3 && tag[2] != "" {
			if err := validateRelayURL(tag[2]); err != nil {
				return fmt.Errorf("invalid participant relay hint: %w", err)
			}
		}

		// Optional role validation (separate field)
		if len(tag) == 4 && tag[3] != "" {
			role := tag[3]
			if len(role) > 100 {
				return fmt.Errorf("participant role too long (max 100 characters)")
			}
		}
	}

	return nil
}

// validateCalendarHashtagTag validates the t (hashtag) tag for calendar events
func validateCalendarHashtagTag(tag nostr.Tag) error {
	if len(tag) != 2 {
		return fmt.Errorf("t tag must have exactly 2 elements")
	}

	hashtag := tag[1]
	if hashtag == "" {
		return fmt.Errorf("hashtag cannot be empty")
	}

	if len(hashtag) > 100 {
		return fmt.Errorf("hashtag too long (max 100 characters)")
	}

	// Hashtags should not contain spaces or special characters
	if strings.ContainsAny(hashtag, " \t\n\r") {
		return fmt.Errorf("hashtag cannot contain whitespace")
	}

	return nil
}

// validateCalendarReferenceTag validates the r (reference) tag for calendar events
func validateCalendarReferenceTag(tag nostr.Tag) error {
	if len(tag) != 2 {
		return fmt.Errorf("r tag must have exactly 2 elements")
	}

	referenceURL := tag[1]
	if err := validateReferenceURL(referenceURL); err != nil {
		return fmt.Errorf("invalid reference URL: %w", err)
	}

	return nil
}

// validateCalendarATag validates the a tag for calendar events (calendar inclusion)
func validateCalendarATag(tag nostr.Tag) error {
	if len(tag) < 2 || len(tag) > 3 {
		return fmt.Errorf("a tag must have 2 or 3 elements")
	}

	aTagValue := tag[1]
	if err := validateCalendarReference(aTagValue); err != nil {
		return fmt.Errorf("invalid calendar reference: %w", err)
	}

	// Optional relay hint validation
	if len(tag) == 3 && tag[2] != "" {
		if err := validateRelayURL(tag[2]); err != nil {
			return fmt.Errorf("invalid calendar reference relay hint: %w", err)
		}
	}

	return nil
}

// validateCalendarNameTag validates the deprecated name tag for calendar events
func validateCalendarNameTag(tag nostr.Tag) error {
	if len(tag) != 2 {
		return fmt.Errorf("name tag must have exactly 2 elements")
	}

	name := tag[1]
	if len(name) > 500 {
		return fmt.Errorf("name too long (max 500 characters)")
	}

	return nil
}

// validateDateTag validates date tags (start/end) for date-based calendar events
func validateDateTag(tag nostr.Tag, tagName string) error {
	if len(tag) != 2 {
		return fmt.Errorf("%s tag must have exactly 2 elements", tagName)
	}

	dateStr := tag[1]
	if dateStr == "" {
		return fmt.Errorf("%s date cannot be empty", tagName)
	}

	// Validate ISO 8601 date format (YYYY-MM-DD)
	if !regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`).MatchString(dateStr) {
		return fmt.Errorf("%s date must be in ISO 8601 format (YYYY-MM-DD)", tagName)
	}

	// Parse date to validate it's actually valid
	if _, err := time.Parse("2006-01-02", dateStr); err != nil {
		return fmt.Errorf("invalid %s date: %w", tagName, err)
	}

	return nil
}

// validateTimestampTag validates timestamp tags (start/end) for time-based calendar events
func validateTimestampTag(tag nostr.Tag, tagName string) error {
	if len(tag) != 2 {
		return fmt.Errorf("%s tag must have exactly 2 elements", tagName)
	}

	timestampStr := tag[1]
	if timestampStr == "" {
		return fmt.Errorf("%s timestamp cannot be empty", tagName)
	}

	// Parse timestamp
	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid %s timestamp: %w", tagName, err)
	}

	// Validate timestamp is reasonable (not negative, not too far in future)
	if timestamp < 0 {
		return fmt.Errorf("%s timestamp cannot be negative", tagName)
	}

	// Don't allow timestamps too far in the future (10 years)
	maxFutureTime := time.Now().AddDate(10, 0, 0).Unix()
	if timestamp > maxFutureTime {
		return fmt.Errorf("%s timestamp is too far in the future", tagName)
	}

	return nil
}

// validateTimezoneTag validates timezone tags (start_tzid/end_tzid) for time-based calendar events
func validateTimezoneTag(tag nostr.Tag, tagName string) error {
	if len(tag) != 2 {
		return fmt.Errorf("%s tag must have exactly 2 elements", tagName)
	}

	timezone := tag[1]
	if timezone == "" {
		return fmt.Errorf("%s timezone cannot be empty", tagName)
	}

	// Validate timezone format (IANA Time Zone Database format)
	// More strict validation for common patterns like America/New_York, Europe/London, etc.
	if !regexp.MustCompile(`^[A-Za-z_]+/[A-Za-z_]+$`).MatchString(timezone) {
		return fmt.Errorf("%s timezone must be in IANA format (e.g., America/New_York)", tagName)
	}

	// Check for known valid timezone patterns to reject obviously invalid ones
	parts := strings.Split(timezone, "/")
	if len(parts) != 2 {
		return fmt.Errorf("%s timezone must be in IANA format (e.g., America/New_York)", tagName)
	}

	// Reject timezone names with "Invalid" in them
	if strings.Contains(strings.ToLower(timezone), "invalid") {
		return fmt.Errorf("%s timezone contains invalid characters", tagName)
	}

	if len(timezone) > 50 {
		return fmt.Errorf("%s timezone too long (max 50 characters)", tagName)
	}

	return nil
}

// validateDateOrdering validates that start date comes before end date
func validateDateOrdering(startDate, endDate string) error {
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return fmt.Errorf("invalid start date: %w", err)
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return fmt.Errorf("invalid end date: %w", err)
	}

	if !start.Before(end) {
		return fmt.Errorf("start date must be before end date")
	}

	return nil
}

// validateCalendarEventReferenceTag validates the a tag for calendar event references
func validateCalendarEventReferenceTag(tag nostr.Tag) error {
	if len(tag) < 2 || len(tag) > 3 {
		return fmt.Errorf("a tag must have 2 or 3 elements")
	}

	// The a tag value might contain relay URL after comma, split it
	aTagValue := tag[1]
	parts := strings.Split(aTagValue, ",")
	actualRef := parts[0] // The actual reference is before any comma
	
	if err := validateCalendarEventReference(actualRef); err != nil {
		return fmt.Errorf("invalid calendar event reference: %w", err)
	}

	// If there's a relay URL in the same field (after comma), validate it
	if len(parts) > 1 && parts[1] != "" {
		if err := validateRelayURL(parts[1]); err != nil {
			return fmt.Errorf("invalid calendar event reference relay hint: %w", err)
		}
	}

	// Optional relay hint validation (separate field)
	if len(tag) == 3 && tag[2] != "" {
		if err := validateRelayURL(tag[2]); err != nil {
			return fmt.Errorf("invalid calendar event reference relay hint: %w", err)
		}
	}

	return nil
}

// validateRSVPATag validates the a tag for RSVP events
func validateRSVPATag(tag nostr.Tag) error {
	if len(tag) < 2 || len(tag) > 3 {
		return fmt.Errorf("a tag must have 2 or 3 elements")
	}

	// The a tag value might contain relay URL after comma, split it
	aTagValue := tag[1]
	parts := strings.Split(aTagValue, ",")
	actualRef := parts[0] // The actual reference is before any comma
	
	if err := validateCalendarEventReference(actualRef); err != nil {
		return fmt.Errorf("invalid calendar event reference: %w", err)
	}

	// If there's a relay URL in the same field (after comma), validate it
	if len(parts) > 1 && parts[1] != "" {
		if err := validateRelayURL(parts[1]); err != nil {
			return fmt.Errorf("invalid RSVP reference relay hint: %w", err)
		}
	}

	// Optional relay hint validation (separate field)
	if len(tag) == 3 && tag[2] != "" {
		if err := validateRelayURL(tag[2]); err != nil {
			return fmt.Errorf("invalid RSVP reference relay hint: %w", err)
		}
	}

	return nil
}

// validateRSVPETag validates the e tag for RSVP events
func validateRSVPETag(tag nostr.Tag) error {
	if len(tag) < 2 || len(tag) > 3 {
		return fmt.Errorf("e tag must have 2 or 3 elements")
	}

	// The e tag value might contain relay URL after comma, split it
	eTagValue := tag[1]
	parts := strings.Split(eTagValue, ",")
	eventID := parts[0] // The actual event ID is before any comma
	
	if len(eventID) != 64 {
		return fmt.Errorf("event ID must be 64 hex characters, got %d", len(eventID))
	}

	if !isHexChar64(eventID) {
		return fmt.Errorf("event ID must be valid hex")
	}

	// If there's a relay URL in the same field (after comma), validate it
	if len(parts) > 1 && parts[1] != "" {
		if err := validateRelayURL(parts[1]); err != nil {
			return fmt.Errorf("invalid RSVP event relay hint: %w", err)
		}
	}

	// Optional relay hint validation (separate field)
	if len(tag) == 3 && tag[2] != "" {
		if err := validateRelayURL(tag[2]); err != nil {
			return fmt.Errorf("invalid RSVP event relay hint: %w", err)
		}
	}

	return nil
}

// validateRSVPStatusTag validates the status tag for RSVP events
func validateRSVPStatusTag(tag nostr.Tag) error {
	if len(tag) != 2 {
		return fmt.Errorf("status tag must have exactly 2 elements")
	}

	status := tag[1]
	validStatuses := map[string]bool{
		"accepted":  true,
		"declined":  true,
		"tentative": true,
	}

	if !validStatuses[status] {
		return fmt.Errorf("invalid status: must be 'accepted', 'declined', or 'tentative'")
	}

	return nil
}

// validateRSVPFreebusyTag validates the fb (free/busy) tag for RSVP events
func validateRSVPFreebusyTag(tag nostr.Tag) error {
	if len(tag) != 2 {
		return fmt.Errorf("fb tag must have exactly 2 elements")
	}

	freeBusy := tag[1]
	validValues := map[string]bool{
		"free": true,
		"busy": true,
	}

	if !validValues[freeBusy] {
		return fmt.Errorf("invalid fb value: must be 'free' or 'busy'")
	}

	return nil
}

// validateRSVPPTag validates the p tag for RSVP events
func validateRSVPPTag(tag nostr.Tag) error {
	if len(tag) < 2 || len(tag) > 3 {
		return fmt.Errorf("p tag must have 2 or 3 elements")
	}

	pubkey := tag[1]
	if len(pubkey) != 64 {
		return fmt.Errorf("pubkey must be 64 hex characters, got %d", len(pubkey))
	}

	if !isHexChar64(pubkey) {
		return fmt.Errorf("pubkey must be valid hex")
	}

	// Optional relay hint validation
	if len(tag) == 3 && tag[2] != "" {
		if err := validateRelayURL(tag[2]); err != nil {
			return fmt.Errorf("invalid RSVP pubkey relay hint: %w", err)
		}
	}

	return nil
}

// validateCalendarReference validates the format of a calendar reference (a tag value)
func validateCalendarReference(aTagValue string) error {
	// Format: kind:pubkey:d_tag_value
	parts := strings.Split(aTagValue, ":")
	if len(parts) != 3 {
		return fmt.Errorf("calendar reference must be in format 'kind:pubkey:d_tag_value', got '%s'", aTagValue)
	}

	// Validate kind
	kind, err := strconv.Atoi(parts[0])
	if err != nil {
		return fmt.Errorf("invalid kind in calendar reference: %s", parts[0])
	}
	if kind != 31924 {
		return fmt.Errorf("calendar reference must reference kind 31924, got %d", kind)
	}

	// Validate pubkey
	pubkey := parts[1]
	if len(pubkey) != 64 {
		return fmt.Errorf("pubkey in calendar reference must be 64 hex characters, got %d", len(pubkey))
	}
	if !isHexChar64(pubkey) {
		return fmt.Errorf("pubkey in calendar reference must be valid hex")
	}

	// Validate d tag value
	dTagValue := parts[2]
	if dTagValue == "" {
		return fmt.Errorf("d tag value in calendar reference cannot be empty")
	}
	if len(dTagValue) > 200 {
		return fmt.Errorf("d tag value in calendar reference too long (max 200 characters)")
	}

	return nil
}

// validateCalendarEventReference validates the format of a calendar event reference (a tag value)
func validateCalendarEventReference(aTagValue string) error {
	// Format: kind:pubkey:d_tag_value
	parts := strings.Split(aTagValue, ":")
	if len(parts) != 3 {
		return fmt.Errorf("calendar event reference must be in format 'kind:pubkey:d_tag_value', got '%s'", aTagValue)
	}

	// Validate kind
	kind, err := strconv.Atoi(parts[0])
	if err != nil {
		return fmt.Errorf("invalid kind in calendar event reference: %s", parts[0])
	}
	if kind != 31922 && kind != 31923 {
		return fmt.Errorf("calendar event reference must reference kind 31922 or 31923, got %d", kind)
	}

	// Validate pubkey
	pubkey := parts[1]
	if len(pubkey) != 64 {
		return fmt.Errorf("pubkey in calendar event reference must be 64 hex characters, got %d", len(pubkey))
	}
	if !isHexChar64(pubkey) {
		return fmt.Errorf("pubkey in calendar event reference must be valid hex")
	}

	// Validate d tag value
	dTagValue := parts[2]
	if dTagValue == "" {
		return fmt.Errorf("d tag value in calendar event reference cannot be empty")
	}
	if len(dTagValue) > 200 {
		return fmt.Errorf("d tag value in calendar event reference too long (max 200 characters)")
	}

	return nil
}

// validateReferenceURL validates reference URLs for r tags
func validateReferenceURL(referenceURL string) error {
	if referenceURL == "" {
		return fmt.Errorf("reference URL cannot be empty")
	}

	u, err := url.Parse(referenceURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	// Allow http/https schemes for references
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("reference URL must use http or https scheme, got %s", u.Scheme)
	}

	if u.Host == "" {
		return fmt.Errorf("reference URL must have a host")
	}

	if len(referenceURL) > 2000 {
		return fmt.Errorf("reference URL too long (max 2000 characters)")
	}

	return nil
}
package common

import (
	"fmt"

	"github.com/Shugur-Network/relay/internal/logger"
	nostr "github.com/nbd-wtf/go-nostr"
	"go.uber.org/zap"
)

// ValidationContext contains metadata about the validation being performed
type ValidationContext struct {
	NIPNumber string
	EventKind int
	EventName string
}

// Validator provides a standardized validation framework for NIP implementations
type Validator struct {
	ctx ValidationContext
}

// NewValidator creates a new validator instance for a specific NIP and event type
func NewValidator(nipNumber string, eventKind int, eventName string) *Validator {
	return &Validator{
		ctx: ValidationContext{
			NIPNumber: nipNumber,
			EventKind: eventKind,
			EventName: eventName,
		},
	}
}

// ValidateBasics performs common validation checks that apply to most NIPs:
// - Nil event check
// - Event kind validation
// - Logs validation start
func (v *Validator) ValidateBasics(event *nostr.Event) error {
	if event == nil {
		return fmt.Errorf("event is nil")
	}

	v.logValidationStart(event)

	if event.Kind != v.ctx.EventKind {
		return fmt.Errorf("invalid kind for %s: expected %d, got %d",
			v.ctx.EventName, v.ctx.EventKind, event.Kind)
	}

	return nil
}

// LogSuccess logs successful validation completion
func (v *Validator) LogSuccess(event *nostr.Event) {
	logger.Debug(fmt.Sprintf("NIP-%s: %s validation successful",
		v.ctx.NIPNumber, v.ctx.EventName),
		zap.String("event_id", event.ID))
}

// LogWarning logs a validation warning
func (v *Validator) LogWarning(event *nostr.Event, message string) {
	logger.Warn(fmt.Sprintf("NIP-%s: %s", v.ctx.NIPNumber, message),
		zap.String("event_id", event.ID),
		zap.String("pubkey", event.PubKey))
}

// FormatError creates a standardized error message for the NIP
func (v *Validator) FormatError(message string, args ...interface{}) error {
	if len(args) > 0 {
		message = fmt.Sprintf(message, args...)
	}
	return fmt.Errorf("NIP-%s %s validation failed: %s", v.ctx.NIPNumber, v.ctx.EventName, message)
}

// logValidationStart logs the beginning of validation
func (v *Validator) logValidationStart(event *nostr.Event) {
	logger.Debug(fmt.Sprintf("NIP-%s: Validating %s",
		v.ctx.NIPNumber, v.ctx.EventName),
		zap.String("event_id", event.ID),
		zap.String("pubkey", event.PubKey))
}

// ValidateEventWithCallback performs basic validation and then calls a custom validation function
// This is useful for NIPs that need additional specific validation logic
func (v *Validator) ValidateEventWithCallback(event *nostr.Event, customValidation func(*nostr.Event) error) error {
	if err := v.ValidateBasics(event); err != nil {
		return err
	}

	if customValidation != nil {
		if err := customValidation(event); err != nil {
			return v.FormatError("%v", err)
		}
	}

	v.LogSuccess(event)
	return nil
}

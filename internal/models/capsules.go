package models

import (
	"time"

	"github.com/nbd-wtf/go-nostr"
)

// TimeCapsuleEvent represents a time capsule event using the existing events table
// This embeds the standard Nostr event structure and adds helper methods
type TimeCapsuleEvent struct {
	nostr.Event
	
	// Cached parsed values for performance (computed from tags)
	UnlockTime    *time.Time `json:"unlock_time,omitempty"`
	Mode         string     `json:"mode,omitempty"`          // "threshold" or "vdf"
	Threshold    int        `json:"threshold,omitempty"`     // t value
	WitnessCount int        `json:"witness_count,omitempty"` // n value
	Witnesses    []string   `json:"witnesses,omitempty"`     // witness npubs
	EncMethod    string     `json:"enc_method,omitempty"`    // encryption method
	Location     string     `json:"location,omitempty"`      // "inline" or "https"
	ContentURI   string     `json:"content_uri,omitempty"`   // if external storage
}

// UnlockShareEvent represents a witness unlock share using the existing events table
type UnlockShareEvent struct {
	nostr.Event
	
	// Cached parsed values for performance (computed from tags)
	CapsuleID    string     `json:"capsule_id,omitempty"`    // referenced capsule event ID
	WitnessNpub  string     `json:"witness_npub,omitempty"`  // witness posting the share
	UnlockTime   *time.Time `json:"unlock_time,omitempty"`   // capsule unlock time
	ProofData    string     `json:"proof_data,omitempty"`    // optional auxiliary proof
}

// CapsuleStatusResponse represents the API response for capsule status queries
// This is computed from events table queries, not stored anywhere
type CapsuleStatusResponse struct {
	ID            string   `json:"id"`
	UnlockTime    int64    `json:"unlock_time"`
	Mode          string   `json:"mode"`
	Threshold     int      `json:"threshold"`
	WitnessCount  int      `json:"witness_count"`
	WitnessCommit string   `json:"witness_commit"`
	SharesSeen    int      `json:"shares_seen"`
	WitnessesSeen []string `json:"witnesses_seen"`
	Locked        bool     `json:"locked"`
	Status        string   `json:"status"`
}

// CapsuleFilter represents query filters for finding capsules
type CapsuleFilter struct {
	PubKey       string     `json:"pubkey,omitempty"`
	Status       string     `json:"status,omitempty"`       // "locked", "unlocked", "expired"
	Mode         string     `json:"mode,omitempty"`         // "threshold", "vdf"
	UnlockedOnly bool       `json:"unlocked_only,omitempty"`
	LockedOnly   bool       `json:"locked_only,omitempty"`
	Before       *time.Time `json:"before,omitempty"`
	After        *time.Time `json:"after,omitempty"`
	Limit        int        `json:"limit,omitempty"`
	Offset       int        `json:"offset,omitempty"`
}

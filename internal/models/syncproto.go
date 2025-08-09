package models

import (
	"github.com/nbd-wtf/go-nostr"
)

// SyncMessage is used for both requests and responses.
//
// For requests (Cmd="SyncRequest"):
//   - 'Since' can be a Unix timestamp (0 if you want everything).
//   - 'Kinds' is an array of the event kinds you want to retrieve.
//   - 'Limit' optional max number of events.
//
// For responses (Cmd="SyncChunk" or "SyncDone"):
//   - 'Events' is set when sending a chunk of events.
//   - 'Kinds' is optional, you might just ignore it in the response.
type SyncMessage struct {
	Cmd    string        `json:"cmd"` // "SyncRequest", "SyncChunk", "SyncDone"
	Since  int64         `json:"since,omitempty"`
	Limit  int           `json:"limit,omitempty"`
	Kinds  []int         `json:"kinds,omitempty"`
	Events []nostr.Event `json:"events,omitempty"`
}

package nips

import (
	"encoding/json"
	"net/http"

	"github.com/Shugur-Network/relay/internal/config"
	"github.com/Shugur-Network/relay/internal/constants"
	nip11 "github.com/nbd-wtf/go-nostr/nip11"
)

// CustomRelayInformationDocument extends the standard NIP-11 document with Time Capsules capability
type CustomRelayInformationDocument struct {
	nip11.RelayInformationDocument
	Capsules *CapsulesCapability `json:"capsules,omitempty"`
}

// CapsulesCapability represents the Time Capsules capability as per the NIP specification
type CapsulesCapability struct {
	Version        string   `json:"v"`
	Modes          []string `json:"modes"`
	MaxInlineBytes int      `json:"max_inline_bytes"`
}

// Nip11Handler handles NIP-11 requests
func Nip11Handler(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	baseMetadata := constants.DefaultRelayMetadata(cfg)
	
	// Create custom metadata with Time Capsules capability
	customMetadata := CustomRelayInformationDocument{
		RelayInformationDocument: baseMetadata,
		Capsules: &CapsulesCapability{
			Version:        "1",
			Modes:          []string{"threshold", "scheduled"},
			MaxInlineBytes: constants.DefaultMaxInlineSize,
		},
	}
	
	ServeCustomRelayMetadata(w, customMetadata)
}

// ServeRelayMetadata serves the relay metadata document
func ServeRelayMetadata(w http.ResponseWriter, metadata nip11.RelayInformationDocument) {
	w.Header().Set("Content-Type", "application/nostr+json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if err := json.NewEncoder(w).Encode(metadata); err != nil {
		http.Error(w, "Failed to encode metadata", http.StatusInternalServerError)
		return
	}
}

// ServeCustomRelayMetadata serves the custom relay metadata document with Time Capsules capability
func ServeCustomRelayMetadata(w http.ResponseWriter, metadata CustomRelayInformationDocument) {
	w.Header().Set("Content-Type", "application/nostr+json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if err := json.NewEncoder(w).Encode(metadata); err != nil {
		http.Error(w, "Failed to encode metadata", http.StatusInternalServerError)
		return
	}
}

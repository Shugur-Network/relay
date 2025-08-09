package nips

import (
	"encoding/json"
	"net/http"

	"github.com/Shugur-Network/Relay/internal/config"
	"github.com/Shugur-Network/Relay/internal/constants"
	nip11 "github.com/nbd-wtf/go-nostr/nip11"
)

// Nip11Handler handles NIP-11 requests
func Nip11Handler(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	metadata := constants.DefaultRelayMetadata(cfg)
	ServeRelayMetadata(w, metadata)
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

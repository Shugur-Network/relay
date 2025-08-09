package identity

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

const (
	// RelayIDFileName is the name of the file where relay ID is stored
	RelayIDFileName = "relay_id.key"
	// RelayIDDir is the directory where relay identity files are stored
	RelayIDDir = ".shugur"
)

// RelayIdentity holds the relay's identity information
type RelayIdentity struct {
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key,omitempty"` // Only stored locally
	RelayID    string `json:"relay_id"`              // Human-readable relay ID
}

// GenerateRelayIdentity creates a new relay identity with ed25519 keypair
func GenerateRelayIdentity() (*RelayIdentity, error) {
	// Generate ed25519 keypair
	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate keypair: %w", err)
	}

	// Convert to nostr format
	pubKeyHex := hex.EncodeToString(publicKey)
	privKeyHex := hex.EncodeToString(privateKey)

	// Generate a human-readable relay ID from public key (first 16 chars)
	relayID := fmt.Sprintf("relay-%s", pubKeyHex[:16])

	return &RelayIdentity{
		PublicKey:  pubKeyHex,
		PrivateKey: privKeyHex,
		RelayID:    relayID,
	}, nil
}

// GetOrCreateRelayIdentity loads existing relay identity or creates a new one
func GetOrCreateRelayIdentity() (*RelayIdentity, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	relayDir := filepath.Join(homeDir, RelayIDDir)
	relayIDPath := filepath.Join(relayDir, RelayIDFileName)

	// Check if relay ID file exists
	if _, err := os.Stat(relayIDPath); os.IsNotExist(err) {
		// Generate new identity
		identity, err := GenerateRelayIdentity()
		if err != nil {
			return nil, fmt.Errorf("failed to generate relay identity: %w", err)
		}

		// Save the private key for future use
		if err := saveRelayIdentity(identity, relayIDPath); err != nil {
			return nil, fmt.Errorf("failed to save relay identity: %w", err)
		}

		return identity, nil
	}

	// Load existing identity
	return loadRelayIdentity(relayIDPath)
}

// saveRelayIdentity saves the relay identity to disk
func saveRelayIdentity(identity *RelayIdentity, path string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// For security, we only store the private key as hex
	// The public key can be derived from it
	content := fmt.Sprintf("%s\n", identity.PrivateKey)

	// Write with restricted permissions
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		return fmt.Errorf("failed to write relay ID file: %w", err)
	}

	return nil
}

// loadRelayIdentity loads the relay identity from disk
func loadRelayIdentity(path string) (*RelayIdentity, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read relay ID file: %w", err)
	}

	// Parse private key (remove any whitespace/newlines)
	privKeyHex := string(content)
	// Remove newline if present
	if len(privKeyHex) > 128 {
		privKeyHex = privKeyHex[:128]
	}

	privKeyBytes, err := hex.DecodeString(privKeyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode private key: %w", err)
	}

	// Derive public key from private key
	privateKey := ed25519.PrivateKey(privKeyBytes)
	publicKey := privateKey.Public().(ed25519.PublicKey)
	pubKeyHex := hex.EncodeToString(publicKey)

	// Generate relay ID
	relayID := fmt.Sprintf("relay-%s", pubKeyHex[:16])

	return &RelayIdentity{
		PublicKey:  pubKeyHex,
		PrivateKey: privKeyHex,
		RelayID:    relayID,
	}, nil
}

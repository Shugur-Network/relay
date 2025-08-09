package models

import "time"

type RelayInfo struct {
	Address   string    `json:"address"`    // The user-facing URL or hostname of the node. Example: "wss://myrelay.example.com"
	PeerID    string    `json:"peer_id"`    // The libp2p Peer ID used by the node on the P2P network
	PublicKey string    `json:"public_key"` //  Identity key if you want to track a separate pubkey
	IsActive  bool      `json:"is_active"`  // Whether we believe this relay is online (reachable) at the moment
	IsSynced  bool      `json:"is_synced"`  // Whether we believe this node is “synced” with our chain of events
	LastSeen  time.Time `json:"last_seen"`  // Timestamp of the last time we successfully saw or checked this node

}

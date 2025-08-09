package application

import (
	"github.com/Shugur-Network/Relay/internal/config"
	"github.com/Shugur-Network/Relay/internal/domain"
	"github.com/Shugur-Network/Relay/internal/storage"
)

// DB returns the node's database instance.
func (n *Node) DB() *storage.DB {
	return n.db
}

// Config returns the node's configuration.
func (n *Node) Config() *config.Config {
	return n.config
}

// GetValidator returns the node's plugin validator.
func (n *Node) GetValidator() domain.EventValidator {
	return n.Validator
}

// GetEventProcessor returns the node's event processor.
func (n *Node) GetEventProcessor() *storage.EventProcessor {
	return n.EventProcessor
}

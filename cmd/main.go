package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Shugur-Network/relay/internal/config"
	"github.com/Shugur-Network/relay/internal/logger"
	"go.uber.org/zap"
)

// These variables are set at build time via -ldflags
var (
	version = "dev"     // Set via -X main.version=...
	commit  = "unknown" // Set via -X main.commit=...
	date    = "unknown" // Set via -X main.date=...
)

func main() {
	// Set version in config package from build information
	config.SetVersion(version)

	// Create a top-level context for the application
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle OS signals for graceful shutdown
	go func() {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
		sig := <-signals
		logger.Info("Received termination signal. Shutting down gracefully...", zap.String("signal", sig.String()))
		cancel()
	}()

	// Check if this is a server command that needs to block
	needsBlocking := false
	if len(os.Args) > 1 && os.Args[1] == "start" {
		// Check if help was requested for start command
		helpRequested := false
		for _, arg := range os.Args[2:] {
			if arg == "--help" || arg == "-h" {
				helpRequested = true
				break
			}
		}
		needsBlocking = !helpRequested
	}

	// Start the CLI entry point
	Execute(ctx)

	// Only block for server commands
	if needsBlocking {
		// Block the main goroutine until the context is canceled
		<-ctx.Done()

		// Perform any cleanup (if needed) before exiting
		logger.Info("Node has shut down successfully.")
		time.Sleep(1 * time.Second) // Give time for logs to flush
	}
}

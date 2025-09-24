package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/Shugur-Network/relay/internal/application"
	"github.com/Shugur-Network/relay/internal/config"
	"github.com/Shugur-Network/relay/internal/logger"
	"github.com/Shugur-Network/relay/internal/metrics"
	"go.uber.org/zap"

	"github.com/spf13/cobra"
)

var (
	cfgFile string         // Path to custom config file (optional)
	cfg     *config.Config // Global reference to loaded configuration
)

// rootCmd defines the main CLI command for shugur relay
var rootCmd = &cobra.Command{
	Use:   "relay",
	Short: "Shugur relay is a high-performance Nostr relay server",
	Long:  `High-performance, reliable, scalable Nostr relay for decentralized communication.`,
	Example: `
  relay start --db-host localhost --db-port 26257
  relay start --log-level debug --metrics-port 9090
  relay start --config /path/to/config.yaml`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip config loading for version command
		if cmd.Name() == "version" {
			return nil
		}

		// Load configuration (use nil logger to avoid sync issues)
		var err error
		cfg, err = config.Load(cfgFile, nil)
		if err != nil {
			return fmt.Errorf("failed to load configuration: %v", err)
		}

		// Override config with command line flags if specified
		flags := cmd.Flags()
		if flags.Changed("relay-name") {
			cfg.Relay.Name, _ = flags.GetString("relay-name")
		}
		if flags.Changed("db-host") {
			cfg.Database.Server, _ = flags.GetString("db-host")
		}
		if flags.Changed("db-port") {
			cfg.Database.Port, _ = flags.GetInt("db-port")
		}
		if flags.Changed("metrics-port") {
			portStr, _ := flags.GetString("metrics-port")
			cfg.Metrics.Port, _ = strconv.Atoi(portStr)
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Default behavior: show help when no subcommand is provided
		if err := cmd.Help(); err != nil {
			fmt.Fprintf(os.Stderr, "Error displaying help: %v\n", err)
		}
	},
}

// Execute runs the root command with the provided context
func Execute(ctx context.Context) {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printWelcomeBanner() {
	fmt.Println("  ____  _                              ____      _             ")
	fmt.Println(" / ___|| |__  _   _  __ _ _   _ _ __  |  _ \\ ___| | __ _ _   _ ")
	fmt.Println(" \\___ \\| '_ \\| | | |/ _` | | | | '__| | |_) / _ \\ |/ _` | | | |")
	fmt.Println("  ___) | | | | |_| | (_| | |_| | |    |  _ <  __/ | (_| | |_| |")
	fmt.Println(" |____/|_| |_|\\__,_|\\__, |\\__,_|_|    |_| \\_\\___|_|\\__,_|\\__, |")
	fmt.Println("                    |___/                                |___/ ")
	fmt.Println()
	fmt.Println("Welcome to Shugur Relay - A high-performance Nostr relay server!")
}

// init is automatically called before main(), sets up flags and loads config
func init() {
	// Add persistent flags (inherited by all subcommands)
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "Path to custom config file (optional)")

	// CLI flags for relay configuration
	rootCmd.PersistentFlags().String("relay-name", "", "Name of the relay (max 30 chars)")
	rootCmd.PersistentFlags().String("db-host", "localhost", "CockroachDB host")
	rootCmd.PersistentFlags().IntP("db-port", "", 26257, "CockroachDB port")
	rootCmd.PersistentFlags().String("log-level", "info", "Logging level (debug, info, warn, error, fatal)")
	rootCmd.PersistentFlags().String("log-file", "", "Path to the log file")
	rootCmd.PersistentFlags().String("log-format", "text", "Log output format (text or json)")
	rootCmd.PersistentFlags().String("metrics-port", "8181", "Port for Prometheus metrics server")

	// A simple version subcommand
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print the version number of shugur relay",
		Long:  "Print the version number of shugur relay along with build information",
		Run: func(cmd *cobra.Command, args []string) {
			// Check if detailed flag is provided
			if detailed, _ := cmd.Flags().GetBool("detailed"); detailed {
				fmt.Println(GetFullVersionInfo())
			} else {
				fmt.Println(GetVersionWithPrefix())
			}
		},
	})

	// Add detailed flag to version command
	versionCmd := rootCmd.Commands()[len(rootCmd.Commands())-1]
	versionCmd.Flags().BoolP("detailed", "d", false, "Show detailed version information")

	// Add start subcommand
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start the shugur relay server",
		Long:  "Start the shugur relay server with the specified configuration",
		Run: func(cmd *cobra.Command, args []string) {
			printWelcomeBanner()

			cfgFile, _ = cmd.Flags().GetString("config")
			if cfgFile != "" {
				absPath, err := filepath.Abs(cfgFile)
				if err != nil {
					logger.Error("Failed to resolve absolute path for config", zap.Error(err))
					os.Exit(1)
				}
				cfgFile = absPath
			}
			logger.Info("Using config file", zap.String("config_file", cfgFile))

			// Use the context passed down from main.go
			ctx := cmd.Context()

			// Initialize metrics
			metrics.RegisterMetrics()

			// Initialize the application/relay
			logger.Info("Starting relay...")
			app, err := application.New(ctx, cfg, nil)
			if err != nil {
				logger.Error("Failed to initialize the relay", zap.Error(err))
				os.Exit(1)
			}

			// Start the relay
			if err := app.Start(ctx); err != nil {
				logger.Error("Failed to start the relay", zap.Error(err))
				os.Exit(1)
			}

			logger.Info("Shugur relay started successfully!")
		},
	}

	rootCmd.AddCommand(startCmd)
}

package config

import (
	"bytes"
	_ "embed"
	"fmt"
	"strings"

	"github.com/Shugur-Network/relay/internal/logger"
	validator "github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

//go:embed defaults.yaml
var defaultYAML []byte

// Version is set at runtime from build information
var Version = "dev" // This will be set by the main package during initialization

var validate = validator.New()

// Config holds every sub‑config.
type Config struct {
	General     GeneralConfig     `mapstructure:"general"      validate:"required"`
	Metrics     MetricsConfig     `mapstructure:"metrics"      validate:"required"`
	Logging     LoggingConfig     `mapstructure:"logging"      validate:"required"`
	Relay       RelayConfig       `mapstructure:"relay"        validate:"required"`
	RelayPolicy RelayPolicyConfig `mapstructure:"relay_policy" validate:"required"`
	Database    DatabaseConfig    `mapstructure:"database"     validate:"required"`
	Capsules    CapsulesConfig    `mapstructure:"capsules"     validate:"required"`
}

// Register custom validation rules
func init() {
	validate.RegisterStructValidation(func(sl validator.StructLevel) {
		cfg := sl.Current().Interface().(Config)

		// Validate nested structs
		if err := validate.Struct(cfg.General); err != nil {
			sl.ReportError(cfg.General, "General", "General", "required", "")
		}
		if err := validate.Struct(cfg.Metrics); err != nil {
			sl.ReportError(cfg.Metrics, "Metrics", "Metrics", "required", "")
		}
		if err := validate.Struct(cfg.Logging); err != nil {
			sl.ReportError(cfg.Logging, "Logging", "Logging", "required", "")
		}
		if err := validate.Struct(cfg.Relay); err != nil {
			sl.ReportError(cfg.Relay, "Relay", "Relay", "required", "")
		}
		if err := validate.Struct(cfg.RelayPolicy); err != nil {
			sl.ReportError(cfg.RelayPolicy, "RelayPolicy", "RelayPolicy", "required", "")
		}
		if err := validate.Struct(cfg.Database); err != nil {
			sl.ReportError(cfg.Database, "Database", "Database", "required", "")
		}
		if err := validate.Struct(cfg.Capsules); err != nil {
			sl.ReportError(cfg.Capsules, "Capsules", "Capsules", "required", "")
		}
	}, Config{})
}

/* ------------------------------------------------------------------ *
|  Public API                                                         |
* -------------------------------------------------------------------*/

// SetVersion sets the version from build information
func SetVersion(v string) {
	Version = v
}

// Load merges defaults → file (optional) → env vars, validates, and returns cfg.
func Load(path string, log *zap.Logger) (*Config, error) {
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetEnvPrefix("SHUGUR") // SHUGUR_GENERAL_LISTENING_PORT
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// 1. defaults.yaml (embedded)
	if err := v.ReadConfig(bytes.NewReader(defaultYAML)); err != nil {
		return nil, fmt.Errorf("read defaults: %w", err)
	}

	// 2. optional user file
	if path != "" {
		v.SetConfigFile(path)
		if err := v.MergeInConfig(); err != nil {
			return nil, fmt.Errorf("read config file: %w", err)
		}
	} else {
		// Check for config.yaml in current directory if no path specified
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		if err := v.MergeInConfig(); err != nil {
			// Config file not found is okay, we'll use defaults
			if log != nil {
				log.Info("No config.yaml found, using defaults")
			}
		} else {
			if log != nil {
				log.Info("Loaded config.yaml from current directory")
			}
		}
	}

	// 3. env already merged by AutomaticEnv()

	var cfg Config
	if err := v.UnmarshalExact(&cfg); err != nil { // ← use Exact
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	if err := validate.Struct(cfg); err != nil {
		return nil, fmt.Errorf("struct validation: %w", err)
	}
	// if err := crossValidate(&cfg); err != nil {
	// 	return nil, err
	// }

	if log != nil {
		log.Info("configuration loaded",
			zap.String("version", Version),
		)
	}
	if err := initializeLogger(cfg.Logging); err != nil {
		return nil, fmt.Errorf("initialize logger: %w", err)
	} else {
		if log != nil {
			log.Info("logger initialized",
				zap.String("level", cfg.Logging.Level),
				zap.String("format", cfg.Logging.Format),
				zap.String("file", cfg.Logging.FilePath),
			)
		}
	}
	return &cfg, nil
}

// MustLoad panics on failure (handy in tests / main()).
func MustLoad(path string, log *zap.Logger) *Config {
	cfg, err := Load(path, log)
	if err != nil {
		panic(err)
	}
	return cfg
}

// initializeLogger initializes the logger using the LoggingConfig
func initializeLogger(loggingConfig LoggingConfig) error {
	return logger.Init(
		logger.WithLevel(loggingConfig.Level),
		logger.WithFormat(loggingConfig.Format),
		logger.WithFile(loggingConfig.FilePath),
		logger.WithVersion(Version),
		logger.WithComponent("relay"),
		logger.WithRotation(loggingConfig.MaxSize, loggingConfig.MaxBackups, loggingConfig.MaxAge),
	)
}

/* ------------------------------------------------------------------ *
|  Cross‑field validation                                             |
* -------------------------------------------------------------------*/

// func crossValidate(cfg *Config) error {
// 	if cfg.Database.MinConnections > cfg.Database.MaxConnections {
// 		return fmt.Errorf("min_connections > max_connections")
// 	}
// 	return nil
// }

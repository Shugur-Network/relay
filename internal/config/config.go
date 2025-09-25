package config

import (
	"bytes"
	_ "embed"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strings"
	"time"

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
	// Register custom validators
	registerCustomValidators()
	
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
		
		// Cross-field validation
		performCrossFieldValidation(sl, cfg)
	}, Config{})
}

// registerCustomValidators registers custom validation functions
func registerCustomValidators() {
	// Validate WebSocket address format
	if err := validate.RegisterValidation("wsaddr", func(fl validator.FieldLevel) bool {
		addr := fl.Field().String()
		if addr == "" {
			return false
		}
		
		// Check if it starts with : (port only) or host:port format
		if strings.HasPrefix(addr, ":") {
			// Port only format like ":8080"
			port := addr[1:]
			if port == "" {
				return false
			}
			// Check if port is numeric and in valid range
			if _, err := net.LookupPort("tcp", port); err != nil {
				return false
			}
			return true
		}
		
		// Host:port format
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return false
		}
		
		// Validate port
		if _, err := net.LookupPort("tcp", port); err != nil {
			return false
		}
		
		// Validate host (can be IP, hostname, or empty for all interfaces)
		if host != "" {
			if ip := net.ParseIP(host); ip == nil {
				// Not an IP, check if it's a valid hostname
				if matched, _ := regexp.MatchString(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$`, host); !matched {
					return false
				}
			}
		}
		
		return true
	}); err != nil {
		logger.Error("Failed to register wsaddr validator", zap.Error(err))
	}
	
	// Validate public key is 64-character hex string
	if err := validate.RegisterValidation("pubkey", func(fl validator.FieldLevel) bool {
		key := fl.Field().String()
		if key == "" {
			return true // Optional field
		}
		if len(key) != 64 {
			return false
		}
		matched, _ := regexp.MatchString(`^[a-fA-F0-9]{64}$`, key)
		return matched
	}); err != nil {
		logger.Error("Failed to register pubkey validator", zap.Error(err))
	}
	
	// Validate duration is reasonable (not too short or too long)
	if err := validate.RegisterValidation("reasonable_duration", func(fl validator.FieldLevel) bool {
		duration := fl.Field().Interface().(time.Duration)
		// Should be between 1 second and 24 hours
		return duration >= time.Second && duration <= 24*time.Hour
	}); err != nil {
		logger.Error("Failed to register reasonable_duration validator", zap.Error(err))
	}
	
	// Validate timeout duration (shorter range)
	if err := validate.RegisterValidation("timeout_duration", func(fl validator.FieldLevel) bool {
		duration := fl.Field().Interface().(time.Duration)
		// Should be between 1 second and 1 hour
		return duration >= time.Second && duration <= time.Hour
	}); err != nil {
		logger.Error("Failed to register timeout_duration validator", zap.Error(err))
	}
	
	// Validate log level
	if err := validate.RegisterValidation("log_level", func(fl validator.FieldLevel) bool {
		level := fl.Field().String()
		validLevels := []string{"debug", "info", "warn", "error", "fatal"}
		for _, valid := range validLevels {
			if level == valid {
				return true
			}
		}
		return false
	}); err != nil {
		logger.Error("Failed to register log_level validator", zap.Error(err))
	}
	
	// Validate log format
	if err := validate.RegisterValidation("log_format", func(fl validator.FieldLevel) bool {
		format := fl.Field().String()
		return format == "console" || format == "json"
	}); err != nil {
		logger.Error("Failed to register log_format validator", zap.Error(err))
	}
	
	// Validate buffer size is power of 2 and reasonable
	if err := validate.RegisterValidation("buffer_size", func(fl validator.FieldLevel) bool {
		size := int(fl.Field().Int())
		if size < 1024 || size > 1048576 { // 1KB to 1MB
			return false
		}
		// Check if it's a power of 2
		return size&(size-1) == 0
	}); err != nil {
		logger.Error("Failed to register buffer_size validator", zap.Error(err))
	}
	
	// Validate hostname or IP
	if err := validate.RegisterValidation("host", func(fl validator.FieldLevel) bool {
		host := fl.Field().String()
		if host == "" {
			return false
		}
		
		// Check if it's an IP address
		if ip := net.ParseIP(host); ip != nil {
			return true
		}
		
		// Check if it's a valid hostname
		matched, _ := regexp.MatchString(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$`, host)
		return matched
	}); err != nil {
		logger.Error("Failed to register host validator", zap.Error(err))
	}
}

// performCrossFieldValidation performs validation across multiple fields
func performCrossFieldValidation(sl validator.StructLevel, cfg Config) {
	// Validate that ban threshold is reasonable compared to rate limits
	if cfg.Relay.ThrottlingConfig.RateLimit.Enabled {
		if cfg.Relay.ThrottlingConfig.BanThreshold > cfg.Relay.ThrottlingConfig.RateLimit.MaxEventsPerSecond*5 {
			sl.ReportError(cfg.Relay.ThrottlingConfig.BanThreshold, "BanThreshold", "BanThreshold", "ban_threshold_too_high", "")
		}
	}
	
	// Validate that event cache size is reasonable for max connections
	if cfg.Relay.EventCacheSize < cfg.Relay.ThrottlingConfig.MaxConnections/10 {
		sl.ReportError(cfg.Relay.EventCacheSize, "EventCacheSize", "EventCacheSize", "cache_size_too_small", "")
	}
	
	// Validate that database port is not the same as metrics port
	if cfg.Database.Port == cfg.Metrics.Port {
		sl.ReportError(cfg.Database.Port, "Port", "Port", "port_conflict", "")
	}
	
	// Validate that public URL scheme matches WebSocket address
	if cfg.Relay.PublicURL != "" {
		if parsedURL, err := url.Parse(cfg.Relay.PublicURL); err == nil {
			if parsedURL.Scheme != "ws" && parsedURL.Scheme != "wss" {
				sl.ReportError(cfg.Relay.PublicURL, "PublicURL", "PublicURL", "invalid_websocket_scheme", "")
			}
		}
	}
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
		return nil, formatValidationError(err)
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

// MustLoad loads configuration and returns error instead of panicking (production-safe)
func MustLoad(path string, log *zap.Logger) (*Config, error) {
	return Load(path, log)
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

// formatValidationError converts validator errors into user-friendly messages
func formatValidationError(err error) error {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		var messages []string
		
		for _, fieldError := range validationErrors {
			message := getFieldErrorMessage(fieldError)
			messages = append(messages, message)
		}
		
		return fmt.Errorf("configuration validation failed:\n  - %s", strings.Join(messages, "\n  - "))
	}
	
	return fmt.Errorf("configuration validation failed: %w", err)
}

// getFieldErrorMessage returns a user-friendly error message for a field validation error
func getFieldErrorMessage(fe validator.FieldError) string {
	field := fe.Field()
	value := fe.Value()
	tag := fe.Tag()
	param := fe.Param()
	
	switch tag {
	case "required":
		return fmt.Sprintf("%s is required but not provided", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s (got: %v)", field, param, value)
	case "max":
		return fmt.Sprintf("%s must be at most %s (got: %v)", field, param, value)
	case "email":
		return fmt.Sprintf("%s must be a valid email address (got: %v)", field, value)
	case "url":
		return fmt.Sprintf("%s must be a valid URL (got: %v)", field, value)
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters long (got: %d)", field, param, len(fmt.Sprintf("%v", value)))
	case "hexadecimal":
		return fmt.Sprintf("%s must contain only hexadecimal characters (got: %v)", field, value)
	case "wsaddr":
		return fmt.Sprintf("%s must be a valid WebSocket address in format ':port' or 'host:port' (got: %v)", field, value)
	case "pubkey":
		return fmt.Sprintf("%s must be a 64-character hexadecimal string (got: %v)", field, value)
	case "reasonable_duration":
		return fmt.Sprintf("%s must be between 1 second and 24 hours (got: %v)", field, value)
	case "timeout_duration":
		return fmt.Sprintf("%s must be between 1 second and 1 hour (got: %v)", field, value)
	case "log_level":
		return fmt.Sprintf("%s must be one of: debug, info, warn, error, fatal (got: %v)", field, value)
	case "log_format":
		return fmt.Sprintf("%s must be either 'console' or 'json' (got: %v)", field, value)
	case "buffer_size":
		return fmt.Sprintf("%s must be a power of 2 between 1KB and 1MB (got: %v)", field, value)
	case "host":
		return fmt.Sprintf("%s must be a valid hostname or IP address (got: %v)", field, value)
	case "ban_threshold_too_high":
		return fmt.Sprintf("%s is too high compared to rate limit settings, should be at most 5x max events per second", field)
	case "cache_size_too_small":
		return fmt.Sprintf("%s is too small for the number of max connections, should be at least 1/10th of max connections", field)
	case "shutdown_timeout_too_short":
		return fmt.Sprintf("%s should be longer than write timeout to allow proper connection closure", field)
	case "port_conflict":
		return "database port conflicts with metrics port, they must be different"
	case "invalid_websocket_scheme":
		return fmt.Sprintf("%s must use 'ws://' or 'wss://' scheme for WebSocket connections", field)
	default:
		return fmt.Sprintf("%s validation failed: %s (got: %v)", field, tag, value)
	}
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

package config

import "time"

// RelayConfig holds relay-specific settings.
type RelayConfig struct {
	Name             string           `mapstructure:"NAME"              json:"name"              validate:"required,min=1,max=30"`
	Description      string           `mapstructure:"DESCRIPTION"       json:"description"       validate:"omitempty,max=200"`
	Contact          string           `mapstructure:"CONTACT"           json:"contact"           validate:"omitempty,email"`
	PublicKey        string           `mapstructure:"PUBLIC_KEY"        json:"public_key"        validate:"omitempty,pubkey"`
	Icon             string           `mapstructure:"ICON"              json:"icon"              validate:"omitempty,url"`
	Banner           string           `mapstructure:"BANNER"            json:"banner"            validate:"omitempty,url"`
	WSAddr           string           `mapstructure:"WS_ADDR"           json:"ws_addr"           validate:"required,wsaddr"`
	PublicURL        string           `mapstructure:"PUBLIC_URL"        json:"public_url"        validate:"omitempty,url"`
	IdleTimeout      time.Duration    `mapstructure:"IDLE_TIMEOUT"      json:"idle_timeout"      validate:"required,reasonable_duration"`
	WriteTimeout     time.Duration    `mapstructure:"WRITE_TIMEOUT"     json:"write_timeout"     validate:"required,timeout_duration"`
	SendBufferSize   int              `mapstructure:"SEND_BUFFER_SIZE"  json:"send_buffer_size"  validate:"required,buffer_size"`
	EventCacheSize   int              `mapstructure:"EVENT_CACHE_SIZE"  json:"event_cache_size"  validate:"required,min=100,max=1000000"`
	ThrottlingConfig ThrottlingConfig `mapstructure:"THROTTLING"        json:"throttling"        validate:"required"`
}

// ThrottlingConfig holds rate limiting settings.
type ThrottlingConfig struct {
	RateLimit      RateLimitConfig `mapstructure:"RATE_LIMIT"         json:"rate_limit"`
	MaxContentLen  int             `mapstructure:"MAX_CONTENT_LENGTH" json:"max_content_length" validate:"required,min=100,max=65536"`
	MaxConnections int             `mapstructure:"MAX_CONNECTIONS"    json:"max_connections"    validate:"required,min=1,max=100000"`
	BanThreshold   int             `mapstructure:"BAN_THRESHOLD"      json:"ban_threshold"      validate:"required,min=1,max=1000"`
	BanDuration    int             `mapstructure:"BAN_DURATION"       json:"ban_duration"       validate:"required,min=1,max=86400"`
}

// RateLimitConfig holds rate limiting settings.
type RateLimitConfig struct {
	Enabled              bool          `mapstructure:"ENABLED"               json:"enabled"`
	MaxEventsPerSecond   int           `mapstructure:"MAX_EVENTS_PER_SECOND" json:"max_events_per_second"   validate:"min=0,max=10000"`
	MaxRequestsPerSecond int           `mapstructure:"MAX_REQUESTS_PER_SECOND" json:"max_requests_per_second" validate:"min=0,max=50000"`
	BurstSize            int           `mapstructure:"BURST_SIZE"            json:"burst_size"              validate:"min=0,max=1000"`
	BanThreshold         int           `mapstructure:"BAN_THRESHOLD"         json:"ban_threshold"           validate:"min=0,max=1000"`
	ProgressiveBan       bool          `mapstructure:"PROGRESSIVE_BAN"       json:"progressive_ban"`
	BanDuration          time.Duration `mapstructure:"BAN_DURATION"          json:"ban_duration"            validate:"reasonable_duration"`
	MaxBanDuration       time.Duration `mapstructure:"MAX_BAN_DURATION"      json:"max_ban_duration"        validate:"reasonable_duration"`
}

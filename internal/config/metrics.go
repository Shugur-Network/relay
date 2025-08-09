package config

// MetricsConfig holds metrics configuration settings.
type MetricsConfig struct {
	Enabled bool `mapstructure:"ENABLED" json:"enabled" validate:"required"`
	Port    int  `mapstructure:"PORT"    json:"port"    validate:"required,min=1024,max=65535"`
}

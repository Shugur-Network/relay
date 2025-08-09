package config

// DatabaseConfig holds database-related settings.
type DatabaseConfig struct {
	// Connection settings
	Server string `mapstructure:"SERVER"            json:"server"            validate:"required"`
	Port   int    `mapstructure:"PORT"             json:"port"             validate:"required,min=1,max=65535"`
}

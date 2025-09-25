package config

// LoggingConfig holds logging-related settings.
type LoggingConfig struct {
	Level      string `mapstructure:"LEVEL"       json:"level"       validate:"required,log_level"`
	FilePath   string `mapstructure:"FILE"        json:"file"        validate:"omitempty"`
	Format     string `mapstructure:"FORMAT"      json:"format"      validate:"omitempty,log_format"`
	MaxSize    int    `mapstructure:"MAX_SIZE"    json:"max_size"    validate:"required,min=1,max=1000"`
	MaxBackups int    `mapstructure:"MAX_BACKUPS" json:"max_backups" validate:"required,min=0,max=100"`
	MaxAge     int    `mapstructure:"MAX_AGE"     json:"max_age"     validate:"required,min=1,max=365"`
}

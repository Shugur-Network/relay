package config

// CapsulesConfig holds time capsules feature settings
type CapsulesConfig struct {
	Enabled      bool `mapstructure:"ENABLED"       json:"enabled"`
	MaxWitnesses int  `mapstructure:"MAX_WITNESSES" json:"max_witnesses" validate:"required,min=1,max=20"`
}

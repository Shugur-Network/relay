package config

// RelayPolicyConfig holds policy settings.
type RelayPolicyConfig struct {
	Blacklist struct {
		PubKeys []string `mapstructure:"PUBKEYS" json:"pubkeys" validate:"omitempty,dive,hexadecimal,len=64"`
	} `mapstructure:"BLACKLIST"`
	Whitelist struct {
		PubKeys []string `mapstructure:"PUBKEYS" json:"pubkeys" validate:"omitempty,dive,hexadecimal,len=64"`
	} `mapstructure:"WHITELIST"`
}

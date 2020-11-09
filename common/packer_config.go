package common

// PackerConfig is a struct that contains the configuration keys that
// are sent by packer, properly tagged already so mapstructure can load
// them. Embed this structure into your configuration class to get it.
type PackerConfig struct {
	PackerBuildName     string            `mapstructure:"packer_build_name"`
	PackerBuilderType   string            `mapstructure:"packer_builder_type"`
	PackerCoreVersion   bool              `mapstructure:"packer_core_version"`
	PackerDebug         bool              `mapstructure:"packer_debug"`
	PackerForce         bool              `mapstructure:"packer_force"`
	PackerOnError       string            `mapstructure:"packer_on_error"`
	PackerUserVars      map[string]string `mapstructure:"packer_user_variables"`
	PackerSensitiveVars []string          `mapstructure:"packer_sensitive_variables"`
}

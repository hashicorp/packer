package common

const (
	// This is the key in configurations that is set to the name of the
	// build.
	BuildNameConfigKey = "packer_build_name"

	// This is the key in the configuration that is set to the type
	// of the builder that is run. This is useful for provisioners and
	// such who want to make use of this.
	BuilderTypeConfigKey = "packer_builder_type"

	// this is the key in the configuration that is set to the version of the
	// Packer Core. This can be used by plugins to set user agents, etc, without
	// having to import the Core to find out the Packer version.
	CoreVersionConfigKey = "packer_core_version"

	// This is the key in configurations that is set to "true" when Packer
	// debugging is enabled.
	DebugConfigKey = "packer_debug"

	// This is the key in configurations that is set to "true" when Packer
	// force build is enabled.
	ForceConfigKey = "packer_force"

	// This key determines what to do when a normal multistep step fails
	// - "cleanup" - run cleanup steps
	// - "abort" - exit without cleanup
	// - "ask" - ask the user
	OnErrorConfigKey = "packer_on_error"

	// TemplatePathKey is the path to the template that configured this build
	TemplatePathKey = "packer_template_path"

	// This key contains a map[string]string of the user variables for
	// template processing.
	UserVariablesConfigKey = "packer_user_variables"
)

// PackerConfig is a struct that contains the configuration keys that
// are sent by packer, properly tagged already so mapstructure can load
// them. Embed this structure into your configuration class to get access to
// this information from the Packer Core.
type PackerConfig struct {
	PackerBuildName     string            `mapstructure:"packer_build_name"`
	PackerBuilderType   string            `mapstructure:"packer_builder_type"`
	PackerCoreVersion   string            `mapstructure:"packer_core_version"`
	PackerDebug         bool              `mapstructure:"packer_debug"`
	PackerForce         bool              `mapstructure:"packer_force"`
	PackerOnError       string            `mapstructure:"packer_on_error"`
	PackerUserVars      map[string]string `mapstructure:"packer_user_variables"`
	PackerSensitiveVars []string          `mapstructure:"packer_sensitive_variables"`
}

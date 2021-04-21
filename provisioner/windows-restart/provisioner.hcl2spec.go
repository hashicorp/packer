// Code generated by "packer-sdc mapstructure-to-hcl2"; DO NOT EDIT.

package restart

import (
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/zclconf/go-cty/cty"
)

// FlatConfig is an auto-generated flat version of Config.
// Where the contents of a field with a `mapstructure:,squash` tag are bubbled up.
type FlatConfig struct {
	PackerBuildName     *string           `mapstructure:"packer_build_name" cty:"packer_build_name" hcl:"packer_build_name"`
	PackerBuilderType   *string           `mapstructure:"packer_builder_type" cty:"packer_builder_type" hcl:"packer_builder_type"`
	PackerCoreVersion   *string           `mapstructure:"packer_core_version" cty:"packer_core_version" hcl:"packer_core_version"`
	PackerDebug         *bool             `mapstructure:"packer_debug" cty:"packer_debug" hcl:"packer_debug"`
	PackerForce         *bool             `mapstructure:"packer_force" cty:"packer_force" hcl:"packer_force"`
	PackerOnError       *string           `mapstructure:"packer_on_error" cty:"packer_on_error" hcl:"packer_on_error"`
	PackerUserVars      map[string]string `mapstructure:"packer_user_variables" cty:"packer_user_variables" hcl:"packer_user_variables"`
	PackerSensitiveVars []string          `mapstructure:"packer_sensitive_variables" cty:"packer_sensitive_variables" hcl:"packer_sensitive_variables"`
	RestartCommand      *string           `mapstructure:"restart_command" cty:"restart_command" hcl:"restart_command"`
	RestartCheckCommand *string           `mapstructure:"restart_check_command" cty:"restart_check_command" hcl:"restart_check_command"`
	RestartTimeout      *string           `mapstructure:"restart_timeout" cty:"restart_timeout" hcl:"restart_timeout"`
	CheckKey            *bool             `mapstructure:"check_registry" cty:"check_registry" hcl:"check_registry"`
	RegistryKeys        []string          `mapstructure:"registry_keys" cty:"registry_keys" hcl:"registry_keys"`
}

// FlatMapstructure returns a new FlatConfig.
// FlatConfig is an auto-generated flat version of Config.
// Where the contents a fields with a `mapstructure:,squash` tag are bubbled up.
func (*Config) FlatMapstructure() interface{ HCL2Spec() map[string]hcldec.Spec } {
	return new(FlatConfig)
}

// HCL2Spec returns the hcl spec of a Config.
// This spec is used by HCL to read the fields of Config.
// The decoded values from this spec will then be applied to a FlatConfig.
func (*FlatConfig) HCL2Spec() map[string]hcldec.Spec {
	s := map[string]hcldec.Spec{
		"packer_build_name":          &hcldec.AttrSpec{Name: "packer_build_name", Type: cty.String, Required: false},
		"packer_builder_type":        &hcldec.AttrSpec{Name: "packer_builder_type", Type: cty.String, Required: false},
		"packer_core_version":        &hcldec.AttrSpec{Name: "packer_core_version", Type: cty.String, Required: false},
		"packer_debug":               &hcldec.AttrSpec{Name: "packer_debug", Type: cty.Bool, Required: false},
		"packer_force":               &hcldec.AttrSpec{Name: "packer_force", Type: cty.Bool, Required: false},
		"packer_on_error":            &hcldec.AttrSpec{Name: "packer_on_error", Type: cty.String, Required: false},
		"packer_user_variables":      &hcldec.AttrSpec{Name: "packer_user_variables", Type: cty.Map(cty.String), Required: false},
		"packer_sensitive_variables": &hcldec.AttrSpec{Name: "packer_sensitive_variables", Type: cty.List(cty.String), Required: false},
		"restart_command":            &hcldec.AttrSpec{Name: "restart_command", Type: cty.String, Required: false},
		"restart_check_command":      &hcldec.AttrSpec{Name: "restart_check_command", Type: cty.String, Required: false},
		"restart_timeout":            &hcldec.AttrSpec{Name: "restart_timeout", Type: cty.String, Required: false},
		"check_registry":             &hcldec.AttrSpec{Name: "check_registry", Type: cty.Bool, Required: false},
		"registry_keys":              &hcldec.AttrSpec{Name: "registry_keys", Type: cty.List(cty.String), Required: false},
	}
	return s
}

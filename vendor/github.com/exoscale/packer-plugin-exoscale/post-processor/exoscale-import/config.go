//go:generate mapstructure-to-hcl2 -type Config

package exoscaleimport

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	pkrconfig "github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

const (
	defaultAPIEndpoint      = "https://api.exoscale.com/v1"
	defaultTemplateBootMode = "legacy"
)

type Config struct {
	SOSEndpoint             string `mapstructure:"sos_endpoint"`
	APIEndpoint             string `mapstructure:"api_endpoint"`
	APIKey                  string `mapstructure:"api_key"`
	APISecret               string `mapstructure:"api_secret"`
	ImageBucket             string `mapstructure:"image_bucket"`
	TemplateZone            string `mapstructure:"template_zone"`
	TemplateName            string `mapstructure:"template_name"`
	TemplateDescription     string `mapstructure:"template_description"`
	TemplateUsername        string `mapstructure:"template_username"`
	TemplateBootMode        string `mapstructure:"template_boot_mode"`
	TemplateDisablePassword bool   `mapstructure:"template_disable_password"`
	TemplateDisableSSHKey   bool   `mapstructure:"template_disable_sshkey"`
	SkipClean               bool   `mapstructure:"skip_clean"`

	ctx interpolate.Context

	common.PackerConfig `mapstructure:",squash"`
}

func NewConfig(raws ...interface{}) (*Config, error) {
	var config Config

	err := pkrconfig.Decode(&config, &pkrconfig.DecodeOpts{
		PluginType:         BuilderId,
		Interpolate:        true,
		InterpolateContext: &config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{},
		},
	}, raws...)
	if err != nil {
		return nil, err
	}

	requiredArgs := map[string]*string{
		"api_key":       &config.APIKey,
		"api_secret":    &config.APISecret,
		"image_bucket":  &config.ImageBucket,
		"template_zone": &config.TemplateZone,
		"template_name": &config.TemplateName,
	}

	errs := new(packer.MultiError)
	for k, v := range requiredArgs {
		if *v == "" {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("%s must be set", k))
		}
	}

	if len(errs.Errors) > 0 {
		return nil, errs
	}

	if config.APIEndpoint == "" {
		config.APIEndpoint = defaultAPIEndpoint
	}

	if config.TemplateBootMode == "" {
		config.TemplateBootMode = defaultTemplateBootMode
	}

	if config.SOSEndpoint == "" {
		config.SOSEndpoint = "https://sos-" + config.TemplateZone + ".exo.io"
	}

	return &config, nil
}

// ConfigSpec returns HCL object spec
func (p *PostProcessor) ConfigSpec() hcldec.ObjectSpec {
	return p.config.FlatMapstructure().HCL2Spec()
}

package classic

import (
	"fmt"
	"net/url"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	Comm                communicator.Config `mapstructure:",squash"`

	Access *AccessConfig

	// Access config overrides
	Username       string `mapstructure:"username"`
	Password       string `mapstructure:"password"`
	IdentityDomain string `mapstructure:"identity_domain"`
	APIEndpoint    string `mapstructure:"api_endpoint"`
	apiEndpointURL *url.URL

	// Image
	ImageName string `mapstructure:"image_name"`
	Shape     string `mapstructure:"shape"`
	ImageList string `mapstructure:"image_list"`
	// Optional. Describes what computers are allowed to reach your instance
	// via SSH. This whitelist must contain the computer you're running Packer
	// from. It defaults to public-internet, meaning that you can SSH into your
	// instance from anywhere as long as you have the right keys
	SSHSourceList string `mapstructure:"ssh_source_list"`

	ctx interpolate.Context
}

func NewConfig(raws ...interface{}) (*Config, error) {
	c := &Config{}

	// Decode from template
	err := config.Decode(c, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &c.ctx,
	}, raws...)
	if err != nil {
		return nil, fmt.Errorf("Failed to mapstructure Config: %+v", err)
	}

	c.apiEndpointURL, err = url.Parse(c.APIEndpoint)
	if err != nil {
		return nil, fmt.Errorf("Error parsing API Endpoint: %s", err)
	}
	// set default source list
	if c.SSHSourceList == "" {
		c.SSHSourceList = "seciplist:/oracle/public/public-internet"
	}

	var errs *packer.MultiError
	if es := c.Comm.Prepare(&c.ctx); len(es) > 0 {
		errs = packer.MultiErrorAppend(errs, es...)
	}

	return c, nil
}

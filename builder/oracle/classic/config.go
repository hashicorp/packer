package classic

import (
	"fmt"
	"net/url"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/config"
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
	ImageList string `json:"image_list"`

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

	return c, nil
}

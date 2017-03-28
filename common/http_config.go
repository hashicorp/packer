package common

import "github.com/mitchellh/packer/template/interpolate"

// HTTPConfig contains configuration for the local HTTP Server
type HTTPConfig struct {
	HTTPDir  string `mapstructure:"http_directory"`
	HTTPPort uint   `mapstructure:"http_port"`
}

func (c *HTTPConfig) Prepare(ctx *interpolate.Context) (err []error) {
	return
}

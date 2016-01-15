package common

import (
	"errors"

	"github.com/mitchellh/packer/template/interpolate"
)

// HTTPConfig contains configuration for the local HTTP Server
type HTTPConfig struct {
	HTTPDir     string `mapstructure:"http_directory"`
	HTTPPortMin uint   `mapstructure:"http_port_min"`
	HTTPPortMax uint   `mapstructure:"http_port_max"`
}

func (c *HTTPConfig) Prepare(ctx *interpolate.Context) []error {
	// Validation
	var errs []error

	if c.HTTPPortMin == 0 {
		c.HTTPPortMin = 8000
	}

	if c.HTTPPortMax == 0 {
		c.HTTPPortMax = 9000
	}

	if c.HTTPPortMin > c.HTTPPortMax {
		errs = append(errs,
			errors.New("http_port_min must be less than http_port_max"))
	}

	return errs
}

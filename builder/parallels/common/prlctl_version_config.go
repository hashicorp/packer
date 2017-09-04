package common

import (
	"github.com/hashicorp/packer/template/interpolate"
)

// PrlctlVersionConfig contains the configuration for `prlctl` version.
type PrlctlVersionConfig struct {
	PrlctlVersionFile string `mapstructure:"prlctl_version_file"`
}

// Prepare sets the default value of "PrlctlVersionFile" property.
func (c *PrlctlVersionConfig) Prepare(ctx *interpolate.Context) []error {
	if c.PrlctlVersionFile == "" {
		c.PrlctlVersionFile = ".prlctl_version"
	}

	return nil
}

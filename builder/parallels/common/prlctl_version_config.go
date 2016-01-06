package common

import (
	"github.com/mitchellh/packer/template/interpolate"
)

type PrlctlVersionConfig struct {
	PrlctlVersionFile string `mapstructure:"prlctl_version_file"`
}

func (c *PrlctlVersionConfig) Prepare(ctx *interpolate.Context) []error {
	if c.PrlctlVersionFile == "" {
		c.PrlctlVersionFile = ".prlctl_version"
	}

	return nil
}

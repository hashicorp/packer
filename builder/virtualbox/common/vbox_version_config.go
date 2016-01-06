package common

import (
	"github.com/mitchellh/packer/template/interpolate"
)

type VBoxVersionConfig struct {
	VBoxVersionFile string `mapstructure:"virtualbox_version_file"`
}

func (c *VBoxVersionConfig) Prepare(ctx *interpolate.Context) []error {
	if c.VBoxVersionFile == "" {
		c.VBoxVersionFile = ".vbox_version"
	}

	return nil
}

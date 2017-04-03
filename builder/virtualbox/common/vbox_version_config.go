package common

import (
	"github.com/mitchellh/packer/template/interpolate"
)

type VBoxVersionConfig struct {
	VBoxVersionFile *string `mapstructure:"virtualbox_version_file"`
}

func (c *VBoxVersionConfig) Prepare(ctx *interpolate.Context) []error {
	if c.VBoxVersionFile == nil {
		default_file := ".vbox_version"
		c.VBoxVersionFile = &default_file
	}

	return nil
}

package common

import (
	"github.com/mitchellh/packer/template/interpolate"
)

type VBoxManagePostConfig struct {
	VBoxManagePost [][]string `mapstructure:"vboxmanage_post"`
}

func (c *VBoxManagePostConfig) Prepare(ctx *interpolate.Context) []error {
	if c.VBoxManagePost == nil {
		c.VBoxManagePost = make([][]string, 0)
	}

	return nil
}

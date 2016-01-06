package common

import (
	"github.com/mitchellh/packer/template/interpolate"
)

type VBoxManageConfig struct {
	VBoxManage [][]string `mapstructure:"vboxmanage"`
}

func (c *VBoxManageConfig) Prepare(ctx *interpolate.Context) []error {
	if c.VBoxManage == nil {
		c.VBoxManage = make([][]string, 0)
	}

	return nil
}

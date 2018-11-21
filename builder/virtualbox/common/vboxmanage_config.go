package common

import (
	"github.com/hashicorp/packer/template/interpolate"
)

type VBoxManageConfig struct {
	VBoxManage []string `mapstructure:"vboxmanage"`
}

func (c *VBoxManageConfig) Prepare(ctx *interpolate.Context) []error {

	return nil
}

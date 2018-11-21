package common

import (
	"github.com/hashicorp/packer/template/interpolate"
)

type VBoxManagePostConfig struct {
	VBoxManagePost []string `mapstructure:"vboxmanage_post"`
}

func (c *VBoxManagePostConfig) Prepare(ctx *interpolate.Context) []error {

	return nil
}

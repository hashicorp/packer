package common

import (
	"github.com/hashicorp/packer/template/interpolate"
)

type VBoxManagePostConfig struct {
	// Identical to vboxmanage,
    // except that it is run after the virtual machine is shutdown, and before the
    // virtual machine is exported.
	VBoxManagePost [][]string `mapstructure:"vboxmanage_post" required:"false"`
}

func (c *VBoxManagePostConfig) Prepare(ctx *interpolate.Context) []error {
	if c.VBoxManagePost == nil {
		c.VBoxManagePost = make([][]string, 0)
	}

	return nil
}

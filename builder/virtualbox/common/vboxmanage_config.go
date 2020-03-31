//go:generate struct-markdown

package common

import (
	"github.com/hashicorp/packer/template/interpolate"
)

type VBoxManageConfig struct {
	// Custom `VBoxManage` commands to execute in order to further customize
	// the virtual machine being created. The value of this is an array of
	// commands to execute. The commands are executed in the order defined in
	// the template. For each command, the command is defined itself as an
	// array of strings, where each string represents a single argument on the
	// command-line to `VBoxManage` (but excluding `VBoxManage` itself). Each
	// arg is treated as a [configuration
	// template](https://packer.io/docs/templates/engine), where the `Name` variable is
	// replaced with the VM name. More details on how to use `VBoxManage` are
	// below.
	VBoxManage [][]string `mapstructure:"vboxmanage" required:"false"`
	// Identical to vboxmanage,
	// except that it is run after the virtual machine is shutdown, and before the
	// virtual machine is exported.
	VBoxManagePost [][]string `mapstructure:"vboxmanage_post" required:"false"`
}

func (c *VBoxManageConfig) Prepare(ctx *interpolate.Context) []error {
	if c.VBoxManage == nil {
		c.VBoxManage = make([][]string, 0)
	}

	if c.VBoxManagePost == nil {
		c.VBoxManagePost = make([][]string, 0)
	}

	return nil
}

//go:generate struct-markdown

package common

import (
	"github.com/hashicorp/packer/packer-plugin-sdk/template/interpolate"
)

//In order to perform extra customization of the virtual machine, a template can
//define extra calls to `VBoxManage` to perform.
//[VBoxManage](https://www.virtualbox.org/manual/ch09.html) is the command-line
//interface to VirtualBox where you can completely control VirtualBox. It can be
//used to do things such as set RAM, CPUs, etc.
type VBoxManageConfig struct {
	// Custom `VBoxManage` commands to execute in order to further customize
	// the virtual machine being created. The example shown below sets the memory and number of CPUs
	// within the virtual machine:
	//
	// In JSON:
	// ```json
	// "vboxmanage": [
	//    ["modifyvm", "{{.Name}}", "--memory", "1024"],
	//	  ["modifyvm", "{{.Name}}", "--cpus", "2"]
	// ]
	// ```
	//
	// In HCL2:
	// ```hcl
	// vboxmanage = [
	//    ["modifyvm", "{{.Name}}", "--memory", "1024"],
	//    ["modifyvm", "{{.Name}}", "--cpus", "2"],
	// ]
	// ```
	//
	// The value of `vboxmanage` is an array of commands to execute. These commands are
	// executed in the order defined. So in the above example, the memory will be set
	// followed by the CPUs.
	// Each command itself is an array of strings, where each string is an argument to
	// `VBoxManage`. Each argument is treated as a [configuration
	// template](/docs/templates/engine). The only available
	// variable is `Name` which is replaced with the unique name of the VM, which is
	// required for many VBoxManage calls.
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

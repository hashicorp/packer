//go:generate struct-markdown

package common

import (
	"github.com/hashicorp/packer/template/interpolate"
)

// PrlctlConfig contains the configuration for running "prlctl" commands
// before the VM start.
type PrlctlConfig struct {
	// Custom prlctl commands to execute
	// in order to further customize the virtual machine being created. The value
	// of this is an array of commands to execute. The commands are executed in the
	// order defined in the template. For each command, the command is defined
	// itself as an array of strings, where each string represents a single
	// argument on the command-line to prlctl (but excluding prlctl itself).
	// Each arg is treated as a configuration
	// template, where the Name
	// variable is replaced with the VM name. More details on how to use prlctl
	// are below.
	Prlctl [][]string `mapstructure:"prlctl" required:"false"`
}

// Prepare sets the default value of "Prlctl" property.
func (c *PrlctlConfig) Prepare(ctx *interpolate.Context) []error {
	if c.Prlctl == nil {
		c.Prlctl = make([][]string, 0)
	}

	return nil
}

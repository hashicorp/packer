//go:generate struct-markdown

package common

import (
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

// PrlctlPostConfig contains the configuration for running "prlctl" commands
// in the end of artifact build.
type PrlctlPostConfig struct {
	// Identical to prlctl, except
	// that it is run after the virtual machine is shutdown, and before the virtual
	// machine is exported.
	PrlctlPost [][]string `mapstructure:"prlctl_post" required:"false"`
}

// Prepare sets the default value of "PrlctlPost" property.
func (c *PrlctlPostConfig) Prepare(ctx *interpolate.Context) []error {
	if c.PrlctlPost == nil {
		c.PrlctlPost = make([][]string, 0)
	}

	return nil
}

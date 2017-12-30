package common

import (
	"github.com/hashicorp/packer/template/interpolate"
)

// PrlctlPostConfig contains the configuration for running "prlctl" commands
// in the end of artifact build.
type PrlctlPostConfig struct {
	PrlctlPost [][]string `mapstructure:"prlctl_post"`
}

// Prepare sets the default value of "PrlctlPost" property.
func (c *PrlctlPostConfig) Prepare(ctx *interpolate.Context) []error {
	if c.PrlctlPost == nil {
		c.PrlctlPost = make([][]string, 0)
	}

	return nil
}

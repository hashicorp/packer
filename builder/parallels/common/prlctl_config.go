package common

import (
	"github.com/hashicorp/packer/template/interpolate"
)

// PrlctlConfig contains the configuration for running "prlctl" commands
// before the VM start.
type PrlctlConfig struct {
	Prlctl [][]string `mapstructure:"prlctl"`
}

// Prepare sets the default value of "Prlctl" property.
func (c *PrlctlConfig) Prepare(ctx *interpolate.Context) []error {
	if c.Prlctl == nil {
		c.Prlctl = make([][]string, 0)
	}

	return nil
}

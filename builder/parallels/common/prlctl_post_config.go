package common

import (
	"github.com/mitchellh/packer/template/interpolate"
)

type PrlctlPostConfig struct {
	PrlctlPost [][]string `mapstructure:"prlctl_post"`
}

func (c *PrlctlPostConfig) Prepare(ctx *interpolate.Context) []error {
	if c.PrlctlPost == nil {
		c.PrlctlPost = make([][]string, 0)
	}

	return nil
}

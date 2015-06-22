package common

import (
	"fmt"
	"github.com/mitchellh/packer/packer"
)

type PrlctlPostConfig struct {
	PrlctlPost [][]string `mapstructure:"prlctl_post"`
}

func (c *PrlctlPostConfig) Prepare(t *packer.ConfigTemplate) []error {
	if c.PrlctlPost == nil {
		c.PrlctlPost = make([][]string, 0)
	}

	errs := make([]error, 0)
	for i, args := range c.PrlctlPost {
		for j, arg := range args {
			if err := t.Validate(arg); err != nil {
				errs = append(errs,
					fmt.Errorf("Error processing prlctl_post[%d][%d]: %s", i, j, err))
			}
		}
	}

	return errs
}

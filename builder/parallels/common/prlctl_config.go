package common

import (
	"fmt"
	"github.com/mitchellh/packer/packer"
)

type PrlctlConfig struct {
	Prlctl [][]string `mapstructure:"prlctl"`
}

func (c *PrlctlConfig) Prepare(t *packer.ConfigTemplate) []error {
	if c.Prlctl == nil {
		c.Prlctl = make([][]string, 0)
	}

	errs := make([]error, 0)
	for i, args := range c.Prlctl {
		for j, arg := range args {
			if err := t.Validate(arg); err != nil {
				errs = append(errs,
					fmt.Errorf("Error processing prlctl[%d][%d]: %s", i, j, err))
			}
		}
	}

	return errs
}

package common

import (
	"fmt"
	"github.com/mitchellh/packer/packer"
)

type VBoxManagePostConfig struct {
	VBoxManagePost [][]string `mapstructure:"vboxmanage_post"`
}

func (c *VBoxManagePostConfig) Prepare(t *packer.ConfigTemplate) []error {
	if c.VBoxManagePost == nil {
		c.VBoxManagePost = make([][]string, 0)
	}

	errs := make([]error, 0)
	for i, args := range c.VBoxManagePost {
		for j, arg := range args {
			if err := t.Validate(arg); err != nil {
				errs = append(errs,
					fmt.Errorf("Error processing vboxmanage_post[%d][%d]: %s", i, j, err))
			}
		}
	}

	return errs
}

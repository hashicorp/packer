package common

import (
	"fmt"
	"github.com/mitchellh/packer/packer"
)

type VBoxManageConfig struct {
	VBoxManage [][]string `mapstructure:"vboxmanage"`
}

func (c *VBoxManageConfig) Prepare(t *packer.ConfigTemplate) []error {
	if c.VBoxManage == nil {
		c.VBoxManage = make([][]string, 0)
	}

	errs := make([]error, 0)
	for i, args := range c.VBoxManage {
		for j, arg := range args {
			if err := t.Validate(arg); err != nil {
				errs = append(errs,
					fmt.Errorf("Error processing vboxmanage[%d][%d]: %s", i, j, err))
			}
		}
	}

	return errs
}

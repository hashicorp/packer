package iso

import (
	"fmt"
	"github.com/mitchellh/packer/packer"
)

type AddDisksConfig struct {
	NewDisks [][]string `mapstructure:"add_disks"`
}

func (c *AddDisksConfig) Prepare(t *packer.ConfigTemplate) []error {
	if c.NewDisks == nil {
		c.NewDisks = make([][]string, 0)
	}
	errs := make([]error, 0)
	for i, args := range c.NewDisks {
		if len(args) != 3 {
			errs = append(errs, fmt.Errorf("Error processing AddDisk[%d]: Incorrect number of arguments", i))
		}
		for j, arg := range args {
			if err := t.Validate(arg); err != nil {
				errs = append(errs,
					fmt.Errorf("Error processing vdiskmanager[%d][%d]: %s", i, j, err))
			}
		}
	}
	return errs
}

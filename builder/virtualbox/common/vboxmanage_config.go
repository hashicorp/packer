package common

import (
	"fmt"
	"strconv"

	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/template/interpolate"
)

type VBoxManageConfig struct {
	VBoxManage [][]string `mapstructure:"vboxmanage"`
}

func (c *VBoxManageConfig) Prepare(ctx *interpolate.Context) []error {
	if c.VBoxManage == nil {
		c.VBoxManage = make([][]string, 0)
		return nil
	}

	var errs []error
	var err error
	var desiredMem uint64

	for _, cmd := range c.VBoxManage {
		if cmd[2] == "--memory" {
			desiredMem, err = strconv.ParseUint(cmd[3], 10, 64)
			if err != nil {
				errs = append(errs, fmt.Errorf("Error parsing string: %s", err))
			}
		}
	}

	if err = common.AvailableMem(desiredMem); err != nil {
		errs = append(errs, fmt.Errorf("Unavailable Resources: %s", err))
	}

	return errs
}

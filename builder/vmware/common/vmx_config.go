package common

import (
	"fmt"
	"strconv"

	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/template/interpolate"
)

type VMXConfig struct {
	VMXData     map[string]string `mapstructure:"vmx_data"`
	VMXDataPost map[string]string `mapstructure:"vmx_data_post"`
}

func (c *VMXConfig) Prepare(ctx *interpolate.Context) []error {
	var errs []error
	var err error
	var desiredMem uint64

	for k, v := range c.VMXData {
		if k == "memsize" {
			desiredMem, err = strconv.ParseUint(v, 10, 64)
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

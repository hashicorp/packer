package common

import (
	"fmt"

	"github.com/mitchellh/packer/packer"
)

type VMXConfig struct {
	VMXData map[string]string `mapstructure:"vmx_data"`
}

func (c *VMXConfig) Prepare(t *packer.ConfigTemplate) []error {
	errs := make([]error, 0)
	newVMXData := make(map[string]string)
	for k, v := range c.VMXData {
		var err error
		k, err = t.Process(k, nil)
		if err != nil {
			errs = append(errs,
				fmt.Errorf("Error processing VMX data key %s: %s", k, err))
			continue
		}

		v, err = t.Process(v, nil)
		if err != nil {
			errs = append(errs,
				fmt.Errorf("Error processing VMX data value '%s': %s", v, err))
			continue
		}

		newVMXData[k] = v
	}
	c.VMXData = newVMXData

	return errs
}

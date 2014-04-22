package common

import (
	"fmt"
	"github.com/mitchellh/packer/packer"
)

type VrdeConfig struct {
	Vrde        bool   `mapstructure:"vrde"`
        VrdeAddress string `mapstructure:"vrdeaddress"`
        VrdePort    uint   `mapstructure:"vrdeport"`
}

func (c *VrdeConfig) Prepare(t *packer.ConfigTemplate) []error {

        if c.VrdeAddress == "" {
                c.VrdeAddress= "127.0.0.1"
        }

        templates := map[string]*string{
                "vrdeaddress": &c.VrdeAddress,
        }

	errs := make([]error, 0)
	for n, ptr := range templates {
		var err error
		*ptr, err = t.Process(*ptr, nil)
		if err != nil {
			errs = append(errs, fmt.Errorf("Error processing %s: %s", n, err))
		}
	}

	return errs
}

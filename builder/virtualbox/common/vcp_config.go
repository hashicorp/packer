package common

import (
	"fmt"
	"github.com/mitchellh/packer/packer"
)

type VcpConfig struct {
	Vcp        bool   `mapstructure:"vcp"`
	VcpFile    string `mapstructure:"vcpfile"`
}

func (c *VcpConfig) Prepare(t *packer.ConfigTemplate) []error {

        templates := map[string]*string{
                "vcpfile": &c.VcpFile,
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

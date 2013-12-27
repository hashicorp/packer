package common

import (
	"fmt"

	"github.com/mitchellh/packer/packer"
)

type DriverConfig struct {
	FusionAppPath string `mapstructure:"fusion_app_path"`
}

func (c *DriverConfig) Prepare(t *packer.ConfigTemplate) []error {
	if c.FusionAppPath == "" {
		c.FusionAppPath = "/Applications/VMware Fusion.app"
	}

	templates := map[string]*string{
		"fusion_app_path": &c.FusionAppPath,
	}

	var err error
	errs := make([]error, 0)
	for n, ptr := range templates {
		*ptr, err = t.Process(*ptr, nil)
		if err != nil {
			errs = append(errs, fmt.Errorf("Error processing %s: %s", n, err))
		}
	}

	return errs
}

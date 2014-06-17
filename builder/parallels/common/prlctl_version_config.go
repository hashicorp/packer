package common

import (
	"fmt"
	"github.com/mitchellh/packer/packer"
)

type PrlctlVersionConfig struct {
	PrlctlVersionFile string `mapstructure:"prlctl_version_file"`
}

func (c *PrlctlVersionConfig) Prepare(t *packer.ConfigTemplate) []error {
	if c.PrlctlVersionFile == "" {
		c.PrlctlVersionFile = ".prlctl_version"
	}

	templates := map[string]*string{
		"prlctl_version_file": &c.PrlctlVersionFile,
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

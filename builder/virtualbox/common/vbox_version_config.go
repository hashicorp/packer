package common

import (
	"fmt"
	"github.com/mitchellh/packer/packer"
)

type VBoxVersionConfig struct {
	VBoxVersionFile string `mapstructure:"virtualbox_version_file"`
}

func (c *VBoxVersionConfig) Prepare(t *packer.ConfigTemplate) []error {
	if c.VBoxVersionFile == "" {
		c.VBoxVersionFile = ".vbox_version"
	}

	templates := map[string]*string{
		"virtualbox_version_file": &c.VBoxVersionFile,
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

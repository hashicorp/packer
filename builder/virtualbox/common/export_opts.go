package common

import (
	"fmt"
	"github.com/mitchellh/packer/packer"
)

type ExportOpts struct {
	ExportOpts string `mapstructure:"export_opts"`
}

func (c *ExportOpts) Prepare(t *packer.ConfigTemplate) []error {
	templates := map[string]*string{
		"export_opts": &c.ExportOpts,
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

package common

import (
	"fmt"
	"github.com/mitchellh/packer/packer"
)

type ExportOpts struct {
	ExportOpts []string `mapstructure:"export_opts"`
}

func (c *ExportOpts) Prepare(t *packer.ConfigTemplate) []error {
	if c.ExportOpts == nil {
		c.ExportOpts = make([]string, 0)
	}

	errs := make([]error, 0)
	for i, str := range c.ExportOpts {
		var err error
		c.ExportOpts[i], err = t.Process(str, nil)
		if err != nil {
			errs = append(errs, fmt.Errorf("Error processing %s: %s", "export_opts", err))
		}
	}

	return errs
}

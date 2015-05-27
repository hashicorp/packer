package common

import (
	"errors"

	"github.com/mitchellh/packer/template/interpolate"
)

type ExportConfig struct {
	Format string `mapstruture:"format"`
}

func (c *ExportConfig) Prepare(ctx *interpolate.Context) []error {
	if c.Format == "" {
		c.Format = "ovf"
	}

	var errs []error
	if c.Format != "ovf" && c.Format != "ova" {
		errs = append(errs,
			errors.New("invalid format, only 'ovf' or 'ova' are allowed"))
	}

	return errs
}

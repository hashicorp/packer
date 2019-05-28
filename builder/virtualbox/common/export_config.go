package common

import (
	"errors"

	"github.com/hashicorp/packer/template/interpolate"
)

type ExportConfig struct {
	// Either ovf or ova, this specifies the output format
    // of the exported virtual machine. This defaults to ovf.
	Format string `mapstructure:"format" required:"false"`
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

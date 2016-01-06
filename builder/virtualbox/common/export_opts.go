package common

import (
	"github.com/mitchellh/packer/template/interpolate"
)

type ExportOpts struct {
	ExportOpts []string `mapstructure:"export_opts"`
}

func (c *ExportOpts) Prepare(ctx *interpolate.Context) []error {
	if c.ExportOpts == nil {
		c.ExportOpts = make([]string, 0)
	}

	return nil
}

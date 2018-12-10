package common

import (
	"fmt"

	"github.com/hashicorp/packer/template/interpolate"
)

type ExportConfig struct {
	Format         string   `mapstructure:"format"`
	OVFToolOptions []string `mapstructure:"ovftool_options"`
	SkipExport     bool     `mapstructure:"skip_export"`
	KeepRegistered bool     `mapstructure:"keep_registered"`
	SkipCompaction bool     `mapstructure:"skip_compaction"`
}

func (c *ExportConfig) Prepare(ctx *interpolate.Context) []error {
	var errs []error
	if c.Format != "" {
		if !(c.Format == "ova" || c.Format == "ovf" || c.Format == "vmx") {
			errs = append(
				errs, fmt.Errorf("format must be one of ova, ovf, or vmx"))
		}
	}
	return errs
}

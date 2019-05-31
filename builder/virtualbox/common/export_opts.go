//go:generate struct-markdown

package common

import (
	"github.com/hashicorp/packer/template/interpolate"
)

type ExportOpts struct {
	// Additional options to pass to the
    // VBoxManage
    // export. This
    // can be useful for passing product information to include in the resulting
    // appliance file. Packer JSON configuration file example:
	ExportOpts []string `mapstructure:"export_opts" required:"false"`
}

func (c *ExportOpts) Prepare(ctx *interpolate.Context) []error {
	if c.ExportOpts == nil {
		c.ExportOpts = make([]string, 0)
	}

	return nil
}

//go:generate struct-markdown

package common

import (
	"github.com/hashicorp/packer/template/interpolate"
)

// PrlctlVersionConfig contains the configuration for `prlctl` version.
type PrlctlVersionConfig struct {
	// The path within the virtual machine to
    // upload a file that contains the prlctl version that was used to create
    // the machine. This information can be useful for provisioning. By default
    // this is ".prlctl_version", which will generally upload it into the
    // home directory.
	PrlctlVersionFile string `mapstructure:"prlctl_version_file" required:"false"`
}

// Prepare sets the default value of "PrlctlVersionFile" property.
func (c *PrlctlVersionConfig) Prepare(ctx *interpolate.Context) []error {
	if c.PrlctlVersionFile == "" {
		c.PrlctlVersionFile = ".prlctl_version"
	}

	return nil
}

//go:generate struct-markdown

package common

import (
	"fmt"

	"github.com/hashicorp/packer/packer-plugin-sdk/template/interpolate"
)

type VBoxVersionConfig struct {
	Communicator string `mapstructure:"communicator"`
	// The path within the virtual machine to
	// upload a file that contains the VirtualBox version that was used to create
	// the machine. This information can be useful for provisioning. By default
	// this is .vbox_version, which will generally be upload it into the
	// home directory. Set to an empty string to skip uploading this file, which
	// can be useful when using the none communicator.
	VBoxVersionFile *string `mapstructure:"virtualbox_version_file" required:"false"`
}

func (c *VBoxVersionConfig) Prepare(ctx *interpolate.Context) []error {
	var errs []error

	if c.VBoxVersionFile == nil {
		default_file := ".vbox_version"
		c.VBoxVersionFile = &default_file
	}

	if c.Communicator == "none" && *c.VBoxVersionFile != "" {
		errs = append(errs, fmt.Errorf("virtualbox_version_file has to be an "+
			"empty string when communicator = 'none'."))
	}

	return errs
}

package common

import (
	"fmt"

	"github.com/hashicorp/packer/template/interpolate"
)

type VBoxVersionConfig struct {
	Communicator    string  `mapstructure:"communicator"`
	VBoxVersionFile *string `mapstructure:"virtualbox_version_file"`
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

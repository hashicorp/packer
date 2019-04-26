package common

import (
	"fmt"

	"github.com/hashicorp/packer/template/interpolate"
)

// These are the different valid mode values for "guest_additions_mode" which
// determine how guest additions are delivered to the guest.
const (
	GuestAdditionsModeDisable string = "disable"
	GuestAdditionsModeAttach         = "attach"
	GuestAdditionsModeUpload         = "upload"
)

type GuestAdditionsConfig struct {
	Communicator       string `mapstructure:"communicator"`
	GuestAdditionsMode string `mapstructure:"guest_additions_mode"`
}

func (c *GuestAdditionsConfig) Prepare(ctx *interpolate.Context) []error {
	var errs []error

	if c.Communicator == "none" && c.GuestAdditionsMode != "disable" {
		errs = append(errs, fmt.Errorf("guest_additions_mode has to be "+
			"'disable' when communicator = 'none'."))
	}

	return errs
}

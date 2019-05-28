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
	// The method by which guest additions are
    // made available to the guest for installation. Valid options are upload,
    // attach, or disable. If the mode is attach the guest additions ISO will
    // be attached as a CD device to the virtual machine. If the mode is upload
    // the guest additions ISO will be uploaded to the path specified by
    // guest_additions_path. The default value is upload. If disable is used,
    // guest additions won't be downloaded, either.
	GuestAdditionsMode string `mapstructure:"guest_additions_mode" required:"false"`
}

func (c *GuestAdditionsConfig) Prepare(ctx *interpolate.Context) []error {
	var errs []error

	if c.Communicator == "none" && c.GuestAdditionsMode != "disable" {
		errs = append(errs, fmt.Errorf("guest_additions_mode has to be "+
			"'disable' when communicator = 'none'."))
	}

	return errs
}

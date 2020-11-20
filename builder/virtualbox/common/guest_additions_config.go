//go:generate struct-markdown

package common

import (
	"fmt"
	"strings"
)

// These are the different valid mode values for "guest_additions_mode" which
// determine how guest additions are delivered to the guest.
const (
	GuestAdditionsModeDisable string = "disable"
	GuestAdditionsModeAttach         = "attach"
	GuestAdditionsModeUpload         = "upload"
)

type GuestAdditionsConfig struct {
	// The method by which guest additions are
	// made available to the guest for installation. Valid options are `upload`,
	// `attach`, or `disable`. If the mode is `attach` the guest additions ISO will
	// be attached as a CD device to the virtual machine. If the mode is `upload`
	// the guest additions ISO will be uploaded to the path specified by
	// `guest_additions_path`. The default value is `upload`. If `disable` is used,
	// guest additions won't be downloaded, either.
	GuestAdditionsMode string `mapstructure:"guest_additions_mode"`
	// The interface type to use to mount guest additions when
	// guest_additions_mode is set to attach. Will default to the value set in
	// iso_interface, if iso_interface is set. Will default to "ide", if
	// iso_interface is not set. Options are "ide" and "sata".
	GuestAdditionsInterface string `mapstructure:"guest_additions_interface" required:"false"`
	// The path on the guest virtual machine
	//  where the VirtualBox guest additions ISO will be uploaded. By default this
	//  is `VBoxGuestAdditions.iso` which should upload into the login directory of
	//  the user. This is a [configuration
	//  template](/docs/templates/engine) where the `Version`
	//  variable is replaced with the VirtualBox version.
	GuestAdditionsPath string `mapstructure:"guest_additions_path"`
	// The SHA256 checksum of the guest
	//  additions ISO that will be uploaded to the guest VM. By default the
	//  checksums will be downloaded from the VirtualBox website, so this only needs
	//  to be set if you want to be explicit about the checksum.
	GuestAdditionsSHA256 string `mapstructure:"guest_additions_sha256"`
	// The URL of the guest additions ISO
	//  to upload. This can also be a file URL if the ISO is at a local path. By
	//  default, the VirtualBox builder will attempt to find the guest additions ISO
	//  on the local file system. If it is not available locally, the builder will
	//  download the proper guest additions ISO from the internet.
	GuestAdditionsURL string `mapstructure:"guest_additions_url" required:"false"`
}

func (c *GuestAdditionsConfig) Prepare(communicatorType string) []error {
	var errs []error

	if c.GuestAdditionsMode == "" {
		c.GuestAdditionsMode = "upload"
	}

	if c.GuestAdditionsPath == "" {
		c.GuestAdditionsPath = "VBoxGuestAdditions.iso"
	}

	if c.GuestAdditionsSHA256 != "" {
		c.GuestAdditionsSHA256 = strings.ToLower(c.GuestAdditionsSHA256)
	}

	validMode := false
	validModes := []string{
		GuestAdditionsModeDisable,
		GuestAdditionsModeAttach,
		GuestAdditionsModeUpload,
	}

	for _, mode := range validModes {
		if c.GuestAdditionsMode == mode {
			validMode = true
			break
		}
	}

	if !validMode {
		errs = append(errs,
			fmt.Errorf("guest_additions_mode is invalid. Must be one of: %v", validModes))
	}

	if communicatorType == "none" && c.GuestAdditionsMode != "disable" {
		errs = append(errs, fmt.Errorf("guest_additions_mode has to be "+
			"'disable' when communicator = 'none'."))
	}

	return errs
}

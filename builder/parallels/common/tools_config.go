package common

import (
	"errors"
	"fmt"

	"github.com/mitchellh/packer/template/interpolate"
)

// These are the different valid mode values for "parallels_tools_mode" which
// determine how guest additions are delivered to the guest.
const (
	ParallelsToolsModeDisable string = "disable"
	ParallelsToolsModeAttach         = "attach"
	ParallelsToolsModeUpload         = "upload"
)

type ToolsConfig struct {
	ParallelsToolsFlavor    string `mapstructure:"parallels_tools_flavor"`
	ParallelsToolsGuestPath string `mapstructure:"parallels_tools_guest_path"`
	ParallelsToolsMode      string `mapstructure:"parallels_tools_mode"`
}

func (c *ToolsConfig) Prepare(ctx *interpolate.Context) []error {
	if c.ParallelsToolsMode == "" {
		c.ParallelsToolsMode = ParallelsToolsModeUpload
	}

	if c.ParallelsToolsGuestPath == "" {
		c.ParallelsToolsGuestPath = "prl-tools-{{.Flavor}}.iso"
	}

	validMode := false
	validModes := []string{
		ParallelsToolsModeDisable,
		ParallelsToolsModeAttach,
		ParallelsToolsModeUpload,
	}

	for _, mode := range validModes {
		if c.ParallelsToolsMode == mode {
			validMode = true
			break
		}
	}

	var errs []error
	if !validMode {
		errs = append(errs,
			fmt.Errorf("parallels_tools_mode is invalid. Must be one of: %v",
				validModes))
	}

	if c.ParallelsToolsFlavor == "" {
		if c.ParallelsToolsMode != ParallelsToolsModeDisable {
			errs = append(errs, errors.New("parallels_tools_flavor must be specified."))
		}
	}

	return errs
}

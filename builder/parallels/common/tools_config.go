package common

import (
	"errors"
	"fmt"

	"github.com/hashicorp/packer/template/interpolate"
)

// These are the different valid mode values for "parallels_tools_mode" which
// determine how guest additions are delivered to the guest.
const (
	ParallelsToolsModeDisable string = "disable"
	ParallelsToolsModeAttach         = "attach"
	ParallelsToolsModeUpload         = "upload"
)

// ToolsConfig contains the builder configuration related to Parallels Tools.
type ToolsConfig struct {
	// The flavor of the Parallels Tools ISO to
    // install into the VM. Valid values are "win", "lin", "mac", "os2"
    // and "other". This can be omitted only if parallels_tools_mode
    // is "disable".
	ParallelsToolsFlavor    string `mapstructure:"parallels_tools_flavor" required:"true"`
	// The path in the virtual machine to
    // upload Parallels Tools. This only takes effect if parallels_tools_mode
    // is "upload". This is a configuration
    // template that has a single
    // valid variable: Flavor, which will be the value of
    // parallels_tools_flavor. By default this is "prl-tools-{{.Flavor}}.iso"
    // which should upload into the login directory of the user.
	ParallelsToolsGuestPath string `mapstructure:"parallels_tools_guest_path" required:"false"`
	// The method by which Parallels Tools are
    // made available to the guest for installation. Valid options are "upload",
    // "attach", or "disable". If the mode is "attach" the Parallels Tools ISO will
    // be attached as a CD device to the virtual machine. If the mode is "upload"
    // the Parallels Tools ISO will be uploaded to the path specified by
    // parallels_tools_guest_path. The default value is "upload".
	ParallelsToolsMode      string `mapstructure:"parallels_tools_mode" required:"false"`
}

// Prepare validates & sets up configuration options related to Parallels Tools.
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
			errs = append(errs, errors.New("parallels_tools_flavor must be specified"))
		}
	}

	return errs
}

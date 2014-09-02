package common

import (
	"errors"
	"fmt"
	"github.com/mitchellh/packer/packer"
	"text/template"
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

func (c *ToolsConfig) Prepare(t *packer.ConfigTemplate) []error {
	if c.ParallelsToolsMode == "" {
		c.ParallelsToolsMode = ParallelsToolsModeUpload
	}

	if c.ParallelsToolsGuestPath == "" {
		c.ParallelsToolsGuestPath = "prl-tools-{{.Flavor}}.iso"
	}

	templates := map[string]*string{
		"parallels_tools_flavor": &c.ParallelsToolsFlavor,
		"parallels_tools_mode":   &c.ParallelsToolsMode,
	}

	var err error
	errs := make([]error, 0)
	for n, ptr := range templates {
		*ptr, err = t.Process(*ptr, nil)
		if err != nil {
			errs = append(errs, fmt.Errorf("Error processing %s: %s", n, err))
		}
	}

	if _, err := template.New("path").Parse(c.ParallelsToolsGuestPath); err != nil {
		errs = append(errs, fmt.Errorf("parallels_tools_guest_path invalid: %s", err))
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

//go:generate struct-markdown
package common

import (
	"fmt"

	"github.com/hashicorp/packer/common/bootcommand"
	"github.com/hashicorp/packer/template/interpolate"
)

// ~> **Note:** For ESXi builds the `disable_vnc` option is ignored. For remote builds, VNC is no longer used to send
// the keystrokes, USB scancodes is used instead.
type VNCConfigWrapper struct {
	bootcommand.VNCConfig `mapstructure:",squash"`
}

func (c *VNCConfigWrapper) Prepare(ctx *interpolate.Context, driverConfig *DriverConfig) (warnings []string, errs []error) {
	if driverConfig.RemoteType != "" {
		if !c.DisableVNC {
			warnings = append(warnings, "[WARN] The vmware-esxi do not use VNC to connect to the host anymore. By default, the VNC is disabled.")
			c.DisableVNC = true
		}
		return
	}

	if len(c.BootCommand) > 0 && c.DisableVNC {
		errs = append(errs,
			fmt.Errorf("A boot command cannot be used when vnc is disabled."))
	}
	errs = append(errs, c.BootConfig.Prepare(ctx)...)
	return
}

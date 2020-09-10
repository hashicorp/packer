//go:generate struct-markdown
package common

import (
	"fmt"

	"github.com/hashicorp/packer/common/bootcommand"
	"github.com/hashicorp/packer/template/interpolate"
)

type BootConfigWrapper struct {
	bootcommand.VNCConfig `mapstructure:",squash"`
	// If set to true, Packer will use USB HID Keyboard scan codes to send the boot command to the VM and
	// the [disable_vnc](#disable_vnc) option will be ignored and automatically set to true.
	// This option is not supported by hosts with free license.
	//
	// ~> **Note:** The ESXi 6.7+ removes support to VNC. In this case, the `usb_keyboard` or [vnc_over_websocket](#vnc_over_websocket)
	// should be set to true in order to send boot command keystrokes to the VM. If both are set, `usb_keyboard` will be ignored
	// and set to false.
	USBKeyBoard bool `mapstructure:"usb_keyboard"`
}

func (c *BootConfigWrapper) Prepare(ctx *interpolate.Context, driverConfig *DriverConfig) (warnings []string, errs []error) {
	if c.USBKeyBoard {
		if driverConfig.RemoteType == "" {
			warnings = append(warnings, "[WARN] `usb_keyboard` can only be used with remote VMWare builds. "+
				"The `usb_keyboard` option will be ignored and automatically set to false.")
			c.USBKeyBoard = false
		} else if !c.DisableVNC {
			warnings = append(warnings, "[WARN] `usb_keyboard` is set to true then the remote VMWare builds "+
				"will not use VNC to connect to the host. The `disable_vnc` option will be ignored and automatically set to true.")
			c.DisableVNC = true
			return
		}
	}

	if len(c.BootCommand) > 0 && c.DisableVNC {
		errs = append(errs,
			fmt.Errorf("A boot command cannot be used when vnc is disabled."))
	}
	errs = append(errs, c.BootConfig.Prepare(ctx)...)
	return
}

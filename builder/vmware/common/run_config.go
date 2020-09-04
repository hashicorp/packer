//go:generate struct-markdown

package common

import (
	"fmt"

	"github.com/hashicorp/packer/template/interpolate"
)

// ~> **Note:** From Packer v1.6.3, remote ESXi builds no longer use VNC to send
// boot_command keystrokes. Any VNC configuration will be ignored.
type RunConfig struct {
	// Packer defaults to building VMware virtual machines
	// by launching a GUI that shows the console of the machine being built. When
	// this value is set to true, the machine will start without a console. For
	// VMware machines, Packer will output VNC connection information in case you
	// need to connect to the console to debug the build process.
	// Some users have experienced issues where Packer cannot properly connect
	// to a VM if it is headless; this appears to be a result of not ever having
	// launched the VMWare GUI and accepting the evaluation license, or
	// supplying a real license. If you experience this, launching VMWare and
	// accepting the license should resolve your problem.
	Headless bool `mapstructure:"headless" required:"false"`
	// The IP address that should be
	// binded to for VNC. By default packer will use 127.0.0.1 for this. If you
	// wish to bind to all interfaces use 0.0.0.0.
	VNCBindAddress string `mapstructure:"vnc_bind_address" required:"false"`
	// The minimum and maximum port
	// to use for VNC access to the virtual machine. The builder uses VNC to type
	// the initial boot_command. Because Packer generally runs in parallel,
	// Packer uses a randomly chosen port in this range that appears available. By
	// default this is 5900 to 6000. The minimum and maximum ports are
	// inclusive.
	VNCPortMin int `mapstructure:"vnc_port_min" required:"false"`
	VNCPortMax int `mapstructure:"vnc_port_max"`
	// Don't auto-generate a VNC password that
	// is used to secure the VNC communication with the VM. This must be set to
	// true if building on ESXi 6.5 and 6.7 with VNC enabled. Defaults to
	// false.
	VNCDisablePassword bool `mapstructure:"vnc_disable_password" required:"false"`
}

func (c *RunConfig) Prepare(ctx *interpolate.Context, driverConfig *DriverConfig) (errs []error) {
	if driverConfig.RemoteType == "" {
		if c.VNCPortMin == 0 {
			c.VNCPortMin = 5900
		}

		if c.VNCPortMax == 0 {
			c.VNCPortMax = 6000
		}

		if c.VNCBindAddress == "" {
			c.VNCBindAddress = "127.0.0.1"
		}

		if c.VNCPortMin > c.VNCPortMax {
			errs = append(errs, fmt.Errorf("vnc_port_min must be less than vnc_port_max"))
		}
		if c.VNCPortMin < 0 {
			errs = append(errs, fmt.Errorf("vnc_port_min must be positive"))
		}
	}
	return
}

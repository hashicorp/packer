//go:generate struct-markdown

package common

import (
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

// ~> **Note:** If [vnc_over_websocket](#vnc_over_websocket) is set to true, any other VNC configuration will be ignored.
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
	// When set to true, Packer will connect to the remote VNC server over a websocket connection
	// and any other VNC configuration option will be ignored.
	// Remote builds using ESXi 6.7+ allows to connect to the VNC server only over websocket,
	// for these the `vnc_over_websocket` must be set to true.
	VNCOverWebsocket bool `mapstructure:"vnc_over_websocket" required:"false"`
	// Do not validate VNC over websocket server's TLS certificate. Defaults to `false`.
	InsecureConnection bool `mapstructure:"insecure_connection" required:"false"`
}

func (c *RunConfig) Prepare(_ *interpolate.Context, driverConfig *DriverConfig) (warnings []string, errs []error) {
	if c.VNCOverWebsocket {
		if driverConfig.RemoteType == "" {
			errs = append(errs, fmt.Errorf("'vnc_over_websocket' can only be used with remote VMWare builds."))
			return
		}
		if c.VNCPortMin != 0 || c.VNCPortMax != 0 || c.VNCBindAddress != "" || c.VNCDisablePassword {
			warnings = append(warnings, "[WARN] When 'vnc_over_websocket' is set "+
				"any other VNC configuration will be ignored.")
		}
		return
	}

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

	return
}

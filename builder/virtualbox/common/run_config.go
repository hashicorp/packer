//go:generate struct-markdown

package common

import (
	"fmt"

	"github.com/hashicorp/packer/template/interpolate"
)

type RunConfig struct {
	// Packer defaults to building VirtualBox virtual
	// machines by launching a GUI that shows the console of the machine
	// being built. When this value is set to true, the machine will start
	// without a console.
	Headless bool `mapstructure:"headless" required:"false"`
	// The IP address that should be
	// binded to for VRDP. By default packer will use 127.0.0.1 for this. If you
	// wish to bind to all interfaces use 0.0.0.0.
	VRDPBindAddress string `mapstructure:"vrdp_bind_address" required:"false"`
	// The minimum and maximum port
	// to use for VRDP access to the virtual machine. Packer uses a randomly chosen
	// port in this range that appears available. By default this is 5900 to
	// 6000. The minimum and maximum ports are inclusive.
	VRDPPortMin int `mapstructure:"vrdp_port_min" required:"false"`
	VRDPPortMax int `mapstructure:"vrdp_port_max"`
}

func (c *RunConfig) Prepare(ctx *interpolate.Context) (errs []error) {
	if c.VRDPBindAddress == "" {
		c.VRDPBindAddress = "127.0.0.1"
	}

	if c.VRDPPortMin == 0 {
		c.VRDPPortMin = 5900
	}

	if c.VRDPPortMax == 0 {
		c.VRDPPortMax = 6000
	}

	if c.VRDPPortMin > c.VRDPPortMax {
		errs = append(
			errs, fmt.Errorf("vrdp_port_min must be less than vrdp_port_max"))
	}

	return
}

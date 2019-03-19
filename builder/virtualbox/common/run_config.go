package common

import (
	"fmt"

	"github.com/hashicorp/packer/template/interpolate"
)

type RunConfig struct {
	Headless bool `mapstructure:"headless"`

	VRDPBindAddress string `mapstructure:"vrdp_bind_address"`
	VRDPPortMin     int    `mapstructure:"vrdp_port_min"`
	VRDPPortMax     int    `mapstructure:"vrdp_port_max"`
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

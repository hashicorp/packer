package common

import (
	"fmt"

	"github.com/hashicorp/packer/template/interpolate"
)

type RunConfig struct {
	Headless bool `mapstructure:"headless"`

	VNCBindAddress     string `mapstructure:"vnc_bind_address"`
	VNCPortMin         int    `mapstructure:"vnc_port_min"`
	VNCPortMax         int    `mapstructure:"vnc_port_max"`
	VNCDisablePassword bool   `mapstructure:"vnc_disable_password"`
}

func (c *RunConfig) Prepare(ctx *interpolate.Context) (errs []error) {
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

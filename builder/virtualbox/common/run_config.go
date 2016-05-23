package common

import (
	"fmt"
	"time"

	"github.com/mitchellh/packer/template/interpolate"
)

type RunConfig struct {
	Headless    bool   `mapstructure:"headless"`
	RawBootWait string `mapstructure:"boot_wait"`

	VRDPBindAddress string `mapstructure:"vrdp_bind_address"`
	VRDPPortMin     uint   `mapstructure:"vrdp_port_min"`
	VRDPPortMax     uint   `mapstructure:"vrdp_port_max"`

	BootWait time.Duration ``
}

func (c *RunConfig) Prepare(ctx *interpolate.Context) []error {
	if c.RawBootWait == "" {
		c.RawBootWait = "10s"
	}

	if c.VRDPBindAddress == "" {
		c.VRDPBindAddress = "127.0.0.1"
	}

	if c.VRDPPortMin == 0 {
		c.VRDPPortMin = 5900
	}

	if c.VRDPPortMax == 0 {
		c.VRDPPortMax = 6000
	}

	var errs []error
	var err error
	c.BootWait, err = time.ParseDuration(c.RawBootWait)
	if err != nil {
		errs = append(errs, fmt.Errorf("Failed parsing boot_wait: %s", err))
	}

	if c.VRDPPortMin > c.VRDPPortMax {
		errs = append(
			errs, fmt.Errorf("vrdp_port_min must be less than vrdp_port_max"))
	}

	return errs
}

package common

import (
	"fmt"
	"time"

	"github.com/mitchellh/packer/template/interpolate"
)

type RunConfig struct {
	Headless    bool   `mapstructure:"headless"`
	RawBootWait string `mapstructure:"boot_wait"`

	VNCBindAddress string `mapstructure:"vnc_bind_address"`
	VNCPortMin     uint   `mapstructure:"vnc_port_min"`
	VNCPortMax     uint   `mapstructure:"vnc_port_max"`

	BootWait time.Duration ``
}

func (c *RunConfig) Prepare(ctx *interpolate.Context) []error {
	if c.RawBootWait == "" {
		c.RawBootWait = "10s"
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

	var errs []error
	var err error
	if c.RawBootWait != "" {
		c.BootWait, err = time.ParseDuration(c.RawBootWait)
		if err != nil {
			errs = append(
				errs, fmt.Errorf("Failed parsing boot_wait: %s", err))
		}
	}

	if c.VNCPortMin > c.VNCPortMax {
		errs = append(
			errs, fmt.Errorf("vnc_port_min must be less than vnc_port_max"))
	}

	return errs
}

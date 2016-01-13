package common

import (
	"errors"
	"fmt"
	"time"

	"github.com/mitchellh/packer/template/interpolate"
)

type RunConfig struct {
	Headless    bool   `mapstructure:"headless"`
	RawBootWait string `mapstructure:"boot_wait"`

	HTTPDir     string `mapstructure:"http_directory"`
	HTTPPortMin uint   `mapstructure:"http_port_min"`
	HTTPPortMax uint   `mapstructure:"http_port_max"`

	VRDPPortMin uint `mapstructure:"vrdp_port_min"`
	VRDPPortMax uint `mapstructure:"vrdp_port_max"`

	BootWait time.Duration ``
}

func (c *RunConfig) Prepare(ctx *interpolate.Context) []error {
	if c.RawBootWait == "" {
		c.RawBootWait = "10s"
	}

	if c.HTTPPortMin == 0 {
		c.HTTPPortMin = 8000
	}

	if c.HTTPPortMax == 0 {
		c.HTTPPortMax = 9000
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

	if c.HTTPPortMin > c.HTTPPortMax {
		errs = append(errs,
			errors.New("http_port_min must be less than http_port_max"))
	}

	if c.VRDPPortMin > c.VRDPPortMax {
		errs = append(
			errs, fmt.Errorf("vrdp_port_min must be less than vrdp_port_max"))
	}

	return errs
}

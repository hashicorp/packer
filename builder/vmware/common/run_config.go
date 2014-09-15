package common

import (
	"errors"
	"fmt"
	"time"

	"github.com/mitchellh/packer/packer"
)

type RunConfig struct {
	Headless    bool   `mapstructure:"headless"`
	RawBootWait string `mapstructure:"boot_wait"`

	HTTPDir     string `mapstructure:"http_directory"`
	HTTPPortMin uint   `mapstructure:"http_port_min"`
	HTTPPortMax uint   `mapstructure:"http_port_max"`

	VNCPortMin uint `mapstructure:"vnc_port_min"`
	VNCPortMax uint `mapstructure:"vnc_port_max"`

	BootWait time.Duration ``
}

func (c *RunConfig) Prepare(t *packer.ConfigTemplate) []error {
	if c.RawBootWait == "" {
		c.RawBootWait = "10s"
	}

	if c.HTTPPortMin == 0 {
		c.HTTPPortMin = 8000
	}

	if c.HTTPPortMax == 0 {
		c.HTTPPortMax = 9000
	}

	if c.VNCPortMin == 0 {
		c.VNCPortMin = 5900
	}

	if c.VNCPortMax == 0 {
		c.VNCPortMax = 6000
	}

	templates := map[string]*string{
		"boot_wait":      &c.RawBootWait,
		"http_directory": &c.HTTPDir,
	}

	var err error
	errs := make([]error, 0)
	for n, ptr := range templates {
		*ptr, err = t.Process(*ptr, nil)
		if err != nil {
			errs = append(errs, fmt.Errorf("Error processing %s: %s", n, err))
		}
	}

	if c.RawBootWait != "" {
		c.BootWait, err = time.ParseDuration(c.RawBootWait)
		if err != nil {
			errs = append(
				errs, fmt.Errorf("Failed parsing boot_wait: %s", err))
		}
	}

	if c.HTTPPortMin > c.HTTPPortMax {
		errs = append(errs,
			errors.New("http_port_min must be less than http_port_max"))
	}

	if c.VNCPortMin > c.VNCPortMax {
		errs = append(
			errs, fmt.Errorf("vnc_port_min must be less than vnc_port_max"))
	}

	return errs
}

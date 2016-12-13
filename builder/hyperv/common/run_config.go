package common

import (
	"fmt"
	"github.com/mitchellh/packer/template/interpolate"
	"time"
)

type RunConfig struct {
	RawBootWait string `mapstructure:"boot_wait"`

	BootWait time.Duration ``
}

func (c *RunConfig) Prepare(ctx *interpolate.Context) []error {
	if c.RawBootWait == "" {
		c.RawBootWait = "10s"
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

	return errs
}

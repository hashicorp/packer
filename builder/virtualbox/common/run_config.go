package common

import (
	"fmt"
	"time"

	"github.com/mitchellh/packer/template/interpolate"
)

type RunConfig struct {
	Headless    bool   `mapstructure:"headless"`
	RawBootWait string `mapstructure:"boot_wait"`

	BootWait time.Duration ``
}

func (c *RunConfig) Prepare(ctx *interpolate.Context) []error {
	if c.RawBootWait == "" {
		c.RawBootWait = "10s"
	}

	var errs []error
	var err error
	c.BootWait, err = time.ParseDuration(c.RawBootWait)
	if err != nil {
		errs = append(errs, fmt.Errorf("Failed parsing boot_wait: %s", err))
	}

	return errs
}

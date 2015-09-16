package common

import (
	"fmt"
	"time"

	"github.com/mitchellh/packer/template/interpolate"
)

type RunConfig struct {
	RawBootWait string `mapstructure:"boot_wait"`

	BootWait time.Duration ``
}

func (c *RunConfig) Prepare(ctx *interpolate.Context) []error {
	if c.RawBootWait == "" {
		c.RawBootWait = "10s"
	}

	var err error
	c.BootWait, err = time.ParseDuration(c.RawBootWait)
	if err != nil {
		return []error{fmt.Errorf("Failed parsing boot_wait: %s", err)}
	}

	return nil
}

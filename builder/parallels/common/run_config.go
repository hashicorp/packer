package common

import (
	"fmt"
	"time"

	"github.com/mitchellh/packer/template/interpolate"
)

// RunConfig contains the configuration for VM run.
type RunConfig struct {
	RawBootWait string `mapstructure:"boot_wait"`

	BootWait time.Duration ``
}

// Prepare sets the configuration for VM run.
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

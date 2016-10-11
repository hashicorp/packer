package common

import (
	"fmt"
	"time"

	"github.com/mitchellh/packer/template/interpolate"
)

type ShutdownConfig struct {
	ShutdownCommand      string `mapstructure:"shutdown_command"`
	RawShutdownTimeout   string `mapstructure:"shutdown_timeout"`
	RawPostShutdownDelay string `mapstructure:"post_shutdown_delay"`

	ShutdownTimeout   time.Duration ``
	PostShutdownDelay time.Duration ``
}

func (c *ShutdownConfig) Prepare(ctx *interpolate.Context) []error {
	if c.RawShutdownTimeout == "" {
		c.RawShutdownTimeout = "5m"
	}

	if c.RawPostShutdownDelay == "" {
		c.RawPostShutdownDelay = "0s"
	}

	var errs []error
	var err error
	c.ShutdownTimeout, err = time.ParseDuration(c.RawShutdownTimeout)
	if err != nil {
		errs = append(errs, fmt.Errorf("Failed parsing shutdown_timeout: %s", err))
	}

	c.PostShutdownDelay, err = time.ParseDuration(c.RawPostShutdownDelay)
	if err != nil {
		errs = append(errs, fmt.Errorf("Failed parsing post_shutdown_delay: %s", err))
	}

	return errs
}

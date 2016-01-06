package common

import (
	"fmt"
	"time"

	"github.com/mitchellh/packer/template/interpolate"
)

type ShutdownConfig struct {
	ShutdownCommand    string `mapstructure:"shutdown_command"`
	RawShutdownTimeout string `mapstructure:"shutdown_timeout"`

	ShutdownTimeout time.Duration ``
}

func (c *ShutdownConfig) Prepare(ctx *interpolate.Context) []error {
	if c.RawShutdownTimeout == "" {
		c.RawShutdownTimeout = "5m"
	}

	var errs []error
	var err error
	c.ShutdownTimeout, err = time.ParseDuration(c.RawShutdownTimeout)
	if err != nil {
		errs = append(errs, fmt.Errorf("Failed parsing shutdown_timeout: %s", err))
	}

	return errs
}

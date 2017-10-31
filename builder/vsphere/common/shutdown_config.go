package common

import (
	"fmt"
	"time"

	"github.com/hashicorp/packer/template/interpolate"
)

type ShutdownConfig struct {
	ShutdownCommand    string `mapstructure:"shutdown_command"`
	RawShutdownTimeout string `mapstructure:"shutdown_timeout"`

	ShutdownTimeout time.Duration ``
}

func (c *ShutdownConfig) Prepare(ctx *interpolate.Context) (warnings []string, errs []error) {
	if c.RawShutdownTimeout == "" {
		c.RawShutdownTimeout = "5m"
	}

	var err error
	c.ShutdownTimeout, err = time.ParseDuration(c.RawShutdownTimeout)
	if err != nil {
		errs = append(errs, fmt.Errorf("Failed parsing shutdown_timeout: %s", err))
	}

	// Warnings
	if c.ShutdownCommand == "" {
		warnings = append(warnings,
			"A shutdown_command was not specified. Without a shutdown command, Packer\n"+
				"will forcibly halt the virtual machine, which may result in data loss.")
	}

	return warnings, errs
}

//go:generate struct-markdown

package common

import (
	"fmt"

	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/template/interpolate"
)

type ShutdownConfig struct {
	// The command to use to gracefully shut down the
	// machine once all the provisioning is done. By default this is an empty
	// string, which tells Packer to just forcefully shut down the machine unless a
	// shutdown command takes place inside script so this may safely be omitted. If
	// one or more scripts require a reboot it is suggested to leave this blank
	// since reboots may fail and specify the final shutdown command in your
	// last script.
	ShutdownCommand string `mapstructure:"shutdown_command" required:"false"`
	// The amount of time to wait after executing the
	// shutdown_command for the virtual machine to actually shut down. If it
	// doesn't shut down in this time, it is an error. By default, the timeout is
	// 5m or five minutes.
	ShutdownTimeout config.DurationString `mapstructure:"shutdown_timeout" required:"false"`
	// The amount of time to wait after shutting
	// down the virtual machine. If you get the error
	// Error removing floppy controller, you might need to set this to 5m
	// or so. By default, the delay is 0s or disabled.
	PostShutdownDelay config.DurationString `mapstructure:"post_shutdown_delay" required:"false"`
}

func (c *ShutdownConfig) Prepare(ctx *interpolate.Context) []error {
	if c.ShutdownTimeout == "" {
		c.ShutdownTimeout = "5m"
	}

	if c.PostShutdownDelay == "" {
		c.PostShutdownDelay = "0s"
	}

	var errs []error
	if err := c.ShutdownTimeout.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("Failed parsing shutdown_timeout: %s", err))
	}

	if err := c.PostShutdownDelay.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("Failed parsing post_shutdown_delay: %s", err))
	}

	return errs
}

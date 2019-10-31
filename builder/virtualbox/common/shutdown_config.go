//go:generate struct-markdown

package common

import (
	"time"

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
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout" required:"false"`
	// The amount of time to wait after shutting
	// down the virtual machine. If you get the error
	// Error removing floppy controller, you might need to set this to 5m
	// or so. By default, the delay is 0s or disabled.
	PostShutdownDelay time.Duration `mapstructure:"post_shutdown_delay" required:"false"`
}

func (c *ShutdownConfig) Prepare(ctx *interpolate.Context) []error {
	if c.ShutdownTimeout == 0 {
		c.ShutdownTimeout = 5 * time.Minute
	}

	return nil
}

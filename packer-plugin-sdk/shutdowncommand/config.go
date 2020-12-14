//go:generate struct-markdown

//Package shutdowncommand is a helper module for builder plugin configuration.
package shutdowncommand

import (
	"time"

	"github.com/hashicorp/packer/packer-plugin-sdk/template/interpolate"
)

// ShutdownConfig defines implementation details for shutting down a VM once it
// is done being provisioned.
//
// It is provided as a convenience to encourage builder developers to
// consider implementing these options, which we believe are valuable for all
// builders. It also helps guarantee that option names for similar options
// are the same across the various builders. Embed it in your builder config
// using the `mapstructure:",squash"` struct tag.
type ShutdownConfig struct {
	// The command to use to gracefully shut down the machine once all
	// provisioning is complete. By default this is an empty string, which
	// tells Packer to just forcefully shut down the machine. This setting can
	// be safely omitted if for example, a shutdown command to gracefully halt
	// the machine is configured inside a provisioning script. If one or more
	// scripts require a reboot it is suggested to leave this blank (since
	// reboots may fail) and instead specify the final shutdown command in your
	// last script.
	ShutdownCommand string `mapstructure:"shutdown_command" required:"false"`
	// The amount of time to wait after executing the shutdown_command for the
	// virtual machine to actually shut down. If the machine doesn't shut down
	// in this time it is considered an error. By default, the time out is "5m"
	// (five minutes).
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout" required:"false"`
}

func (c *ShutdownConfig) Prepare(ctx *interpolate.Context) []error {
	if c.ShutdownTimeout == 0 {
		c.ShutdownTimeout = 5 * time.Minute
	}

	return nil
}

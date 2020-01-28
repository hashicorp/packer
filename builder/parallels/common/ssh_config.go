package common

import (
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/template/interpolate"
	"time"
)

// SSHConfig contains the configuration for SSH communicator.
type SSHConfig struct {
	Comm communicator.Config `mapstructure:",squash"`

	// These are deprecated, but we keep them around for BC
	// TODO: remove later
	SSHWaitTimeout time.Duration `mapstructure:"ssh_wait_timeout" required:"false"`
}

// Prepare sets the default values for SSH communicator properties.
func (c *SSHConfig) Prepare(ctx *interpolate.Context) []error {
	// Backwards compatibility
	if c.SSHWaitTimeout != 0 {
		c.Comm.SSHTimeout = c.SSHWaitTimeout
	}

	return c.Comm.Prepare(ctx)
}

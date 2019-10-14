//go:generate struct-markdown


package common

import (
	"time"

	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/template/interpolate"
)

type SSHConfig struct {
	Comm communicator.Config `mapstructure:",squash"`

	// These are deprecated, but we keep them around for BC
	// TODO(@mitchellh): remove
	SSHSkipRequestPty bool `mapstructure:"ssh_skip_request_pty"`
	// These are deprecated, but we keep them around for BC
	// TODO(@mitchellh): remove
	SSHWaitTimeout time.Duration `mapstructure:"ssh_wait_timeout"`
}

func (c *SSHConfig) Prepare(ctx *interpolate.Context) []error {
	// TODO: backwards compatibility, write fixer instead
	if c.SSHWaitTimeout != 0 {
		c.Comm.SSHTimeout = c.SSHWaitTimeout
	}
	if c.SSHSkipRequestPty {
		c.Comm.SSHPty = false
	}

	return c.Comm.Prepare(ctx)
}

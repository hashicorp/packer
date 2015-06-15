package common

import (
	"time"

	"github.com/mitchellh/packer/helper/communicator"
	"github.com/mitchellh/packer/template/interpolate"
)

type SSHConfig struct {
	Comm communicator.Config `mapstructure:",squash"`

	// These are deprecated, but we keep them around for BC
	// TODO(@mitchellh): remove
	SSHKeyPath     string        `mapstructure:"ssh_key_path"`
	SSHWaitTimeout time.Duration `mapstructure:"ssh_wait_timeout"`
}

func (c *SSHConfig) Prepare(ctx *interpolate.Context) []error {
	// TODO: backwards compatibility, write fixer instead
	if c.SSHKeyPath != "" {
		c.Comm.SSHPrivateKey = c.SSHKeyPath
	}
	if c.SSHWaitTimeout != 0 {
		c.Comm.SSHTimeout = c.SSHWaitTimeout
	}

	return c.Comm.Prepare(ctx)
}

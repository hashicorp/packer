package common

import (
	"errors"
	"time"

	"github.com/mitchellh/packer/helper/communicator"
	"github.com/mitchellh/packer/template/interpolate"
)

type SSHConfig struct {
	Comm communicator.Config `mapstructure:",squash"`

	SSHHostPortMin    uint `mapstructure:"ssh_host_port_min"`
	SSHHostPortMax    uint `mapstructure:"ssh_host_port_max"`
	SSHSkipNatMapping bool `mapstructure:"ssh_skip_nat_mapping"`

	// These are deprecated, but we keep them around for BC
	// TODO(@mitchellh): remove
	SSHKeyPath     string        `mapstructure:"ssh_key_path"`
	SSHWaitTimeout time.Duration `mapstructure:"ssh_wait_timeout"`
}

func (c *SSHConfig) Prepare(ctx *interpolate.Context) []error {
	if c.SSHHostPortMin == 0 {
		c.SSHHostPortMin = 2222
	}

	if c.SSHHostPortMax == 0 {
		c.SSHHostPortMax = 4444
	}

	// TODO: backwards compatibility, write fixer instead
	if c.SSHKeyPath != "" {
		c.Comm.SSHPrivateKey = c.SSHKeyPath
	}
	if c.SSHWaitTimeout != 0 {
		c.Comm.SSHTimeout = c.SSHWaitTimeout
	}

	errs := c.Comm.Prepare(ctx)
	if c.SSHHostPortMin > c.SSHHostPortMax {
		errs = append(errs,
			errors.New("ssh_host_port_min must be less than ssh_host_port_max"))
	}

	return errs
}

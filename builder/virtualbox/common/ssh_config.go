package common

import (
	"errors"
	"time"

	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/template/interpolate"
)

type SSHConfig struct {
	Comm communicator.Config `mapstructure:",squash"`

	SSHHostPortMin    int  `mapstructure:"ssh_host_port_min"`
	SSHHostPortMax    int  `mapstructure:"ssh_host_port_max"`
	SSHSkipNatMapping bool `mapstructure:"ssh_skip_nat_mapping"`

	// These are deprecated, but we keep them around for BC
	// TODO(@mitchellh): remove
	SSHWaitTimeout time.Duration `mapstructure:"ssh_wait_timeout"`
}

func (c *SSHConfig) Prepare(ctx *interpolate.Context) []error {
	if c.Comm.SSHHost == "" {
		c.Comm.SSHHost = "127.0.0.1"
	}

	if c.SSHHostPortMin == 0 {
		c.SSHHostPortMin = 2222
	}

	if c.SSHHostPortMax == 0 {
		c.SSHHostPortMax = 4444
	}

	// TODO: backwards compatibility, write fixer instead
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

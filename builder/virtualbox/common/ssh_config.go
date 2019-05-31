//go:generate struct-markdown

package common

import (
	"errors"
	"time"

	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/template/interpolate"
)

type SSHConfig struct {
	Comm communicator.Config `mapstructure:",squash"`
	// The minimum and
    // maximum port to use for the SSH port on the host machine which is forwarded
    // to the SSH port on the guest machine. Because Packer often runs in parallel,
    // Packer will choose a randomly available port in this range to use as the
    // host port. By default this is 2222 to 4444.
	SSHHostPortMin    int  `mapstructure:"ssh_host_port_min" required:"false"`
	SSHHostPortMax    int  `mapstructure:"ssh_host_port_max"`
	// Defaults to false. When enabled, Packer
    // does not setup forwarded port mapping for SSH requests and uses ssh_port
    // on the host to communicate to the virtual machine.
	SSHSkipNatMapping bool `mapstructure:"ssh_skip_nat_mapping" required:"false"`

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

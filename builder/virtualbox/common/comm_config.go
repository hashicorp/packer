//go:generate struct-markdown

package common

import (
	"errors"

	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/template/interpolate"
)

type CommConfig struct {
	Comm communicator.Config `mapstructure:",squash"`
	// The minimum port to use for the Communicator port on the host machine which is forwarded
	// to the SSH or WinRM port on the guest machine. By default this is 2222.
	HostPortMin int `mapstructure:"host_port_min" required:"false"`
	// The maximum port to use for the Communicator port on the host machine which is forwarded
	// to the SSH or WinRM port on the guest machine. Because Packer often runs in parallel,
	// Packer will choose a randomly available port in this range to use as the
	// host port. By default this is 4444.
	HostPortMax int `mapstructure:"host_port_max" required:"false"`
	// Defaults to false. When enabled, Packer
	// does not setup forwarded port mapping for communicator (SSH or WinRM) requests and uses ssh_port or winrm_port
	// on the host to communicate to the virtual machine.
	SkipNatMapping bool `mapstructure:"skip_nat_mapping" required:"false"`

	// These are deprecated, but we keep them around for backwards compatibility
	// TODO: remove later
	SSHHostPortMin int `mapstructure:"ssh_host_port_min" required:"false"`
	// TODO: remove later
	SSHHostPortMax int `mapstructure:"ssh_host_port_max"`
	// TODO: remove later
	SSHSkipNatMapping bool `mapstructure:"ssh_skip_nat_mapping" required:"false"`
}

func (c *CommConfig) Prepare(ctx *interpolate.Context) []error {
	// Backwards compatibility
	if c.SSHHostPortMin != 0 {
		c.HostPortMin = c.SSHHostPortMin
	}

	// Backwards compatibility
	if c.SSHHostPortMax != 0 {
		c.HostPortMax = c.SSHHostPortMax
	}

	// Backwards compatibility
	if c.SSHSkipNatMapping {
		c.SkipNatMapping = c.SSHSkipNatMapping
	}

	if c.Comm.SSHHost == "" {
		c.Comm.SSHHost = "127.0.0.1"
	}

	if c.HostPortMin == 0 {
		c.HostPortMin = 2222
	}

	if c.HostPortMax == 0 {
		c.HostPortMax = 4444
	}

	errs := c.Comm.Prepare(ctx)
	if c.HostPortMin > c.HostPortMax {
		errs = append(errs,
			errors.New("host_port_min must be less than host_port_max"))
	}

	return errs
}

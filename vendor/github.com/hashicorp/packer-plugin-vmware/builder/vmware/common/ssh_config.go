//go:generate packer-sdc struct-markdown

package common

import (
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

type SSHConfig struct {
	Comm communicator.Config `mapstructure:",squash"`

	// These are deprecated, but we keep them around for BC
	// TODO(@mitchellh): remove
	SSHSkipRequestPty bool `mapstructure:"ssh_skip_request_pty"`
}

func (c *SSHConfig) Prepare(ctx *interpolate.Context) []error {
	if c.SSHSkipRequestPty {
		c.Comm.SSHPty = false
	}

	return c.Comm.Prepare(ctx)
}

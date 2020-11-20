package common

import (
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/interpolate"
)

// SSHConfig contains the configuration for SSH communicator.
type SSHConfig struct {
	Comm communicator.Config `mapstructure:",squash"`
}

// Prepare sets the default values for SSH communicator properties.
func (c *SSHConfig) Prepare(ctx *interpolate.Context) []error {
	return c.Comm.Prepare(ctx)
}

package common

import (
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/template/interpolate"
)

type SSHConfig struct {
	Comm communicator.Config `mapstructure:",squash"`
}

func (c *SSHConfig) Prepare(ctx *interpolate.Context) []error {
	return c.Comm.Prepare(ctx)
}

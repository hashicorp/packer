package common

import (
	"time"

	"github.com/mitchellh/packer/helper/communicator"
	"github.com/mitchellh/packer/template/interpolate"
)

type SSHConfig struct {
	Comm communicator.Config `mapstructure:",squash"`
}

func (c *SSHConfig) Prepare(ctx *interpolate.Context) []error {
	return c.Comm.Prepare(ctx)
}

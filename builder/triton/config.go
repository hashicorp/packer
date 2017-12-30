package triton

import (
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/template/interpolate"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	AccessConfig        `mapstructure:",squash"`
	SourceMachineConfig `mapstructure:",squash"`
	TargetImageConfig   `mapstructure:",squash"`

	Comm communicator.Config `mapstructure:",squash"`

	ctx interpolate.Context
}

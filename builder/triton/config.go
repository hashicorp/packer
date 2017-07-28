package triton

import (
	"github.com/cstuntz/packer/common"
	"github.com/cstuntz/packer/helper/communicator"
	"github.com/cstuntz/packer/template/interpolate"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	AccessConfig        `mapstructure:",squash"`
	SourceMachineConfig `mapstructure:",squash"`
	TargetImageConfig   `mapstructure:",squash"`

	Comm communicator.Config `mapstructure:",squash"`

	ctx interpolate.Context
}

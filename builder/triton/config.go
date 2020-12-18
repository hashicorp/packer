//go:generate mapstructure-to-hcl2 -type Config

package triton

import (
	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	AccessConfig        `mapstructure:",squash"`
	SourceMachineConfig `mapstructure:",squash"`
	TargetImageConfig   `mapstructure:",squash"`

	Comm communicator.Config `mapstructure:",squash"`

	ctx interpolate.Context
}

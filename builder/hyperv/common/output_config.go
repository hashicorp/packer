package common

import (
	"fmt"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/template/interpolate"
)

type OutputConfig struct {
	OutputDir string `mapstructure:"output_directory"`
}

func (c *OutputConfig) Prepare(ctx *interpolate.Context, pc *common.PackerConfig) []error {
	if c.OutputDir == "" {
		c.OutputDir = fmt.Sprintf("output-%s", pc.PackerBuildName)
	}

	return nil
}

//go:generate struct-markdown

package common

import (
	"fmt"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/template/interpolate"
)

type OutputConfig struct {
	// This setting specifies the directory that
    // artifacts from the build, such as the virtual machine files and disks,
    // will be output to. The path to the directory may be relative or
    // absolute. If relative, the path is relative to the working directory
    // packer is executed from. This directory must not exist or, if
    // created, must be empty prior to running the builder. By default this is
    // "output-BUILDNAME" where "BUILDNAME" is the name of the build.
	OutputDir string `mapstructure:"output_directory" required:"false"`
}

func (c *OutputConfig) Prepare(ctx *interpolate.Context, pc *common.PackerConfig) []error {
	if c.OutputDir == "" {
		c.OutputDir = fmt.Sprintf("output-%s", pc.PackerBuildName)
	}

	return nil
}

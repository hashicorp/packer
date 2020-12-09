//go:generate struct-markdown

package common

import (
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

type OutputConfig struct {
	// This is the path to the directory where the
	// resulting virtual machine will be created. This may be relative or absolute.
	// If relative, the path is relative to the working directory when packer
	// is executed. This directory must not exist or be empty prior to running
	// the builder. By default this is output-BUILDNAME where "BUILDNAME" is the
	// name of the build.
	OutputDir string `mapstructure:"output_directory" required:"false"`
	// This is the base name of the file (excluding the file extension) where
	// the resulting virtual machine will be created. By default this is the
	// `vm_name`.
	OutputFilename string `mapstructure:"output_filename" required:"false"`
}

func (c *OutputConfig) Prepare(ctx *interpolate.Context, pc *common.PackerConfig) []error {
	if c.OutputDir == "" {
		c.OutputDir = fmt.Sprintf("output-%s", pc.PackerBuildName)
	}

	return nil
}

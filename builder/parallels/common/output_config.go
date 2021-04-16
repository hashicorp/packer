//go:generate packer-sdc struct-markdown

package common

import (
	"fmt"
	"os"
	"path"

	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

// OutputConfig contains the configuration for builder's output.
type OutputConfig struct {
	// This is the path to the directory where the
	// resulting virtual machine will be created. This may be relative or absolute.
	// If relative, the path is relative to the working directory when packer
	// is executed. This directory must not exist or be empty prior to running
	// the builder. By default this is "output-BUILDNAME" where "BUILDNAME" is the
	// name of the build.
	OutputDir string `mapstructure:"output_directory" required:"false"`
}

// Prepare configures the output directory or returns an error if it already exists.
func (c *OutputConfig) Prepare(ctx *interpolate.Context, pc *common.PackerConfig) []error {
	if c.OutputDir == "" {
		c.OutputDir = fmt.Sprintf("output-%s", pc.PackerBuildName)
	}

	var errs []error

	if path.IsAbs(c.OutputDir) {
		c.OutputDir = path.Clean(c.OutputDir)
	} else {
		wd, err := os.Getwd()
		if err != nil {
			errs = append(errs, err)
		}
		c.OutputDir = path.Clean(path.Join(wd, c.OutputDir))
	}

	if !pc.PackerForce {
		if _, err := os.Stat(c.OutputDir); err == nil {
			errs = append(errs, fmt.Errorf(
				"Output directory '%s' already exists. It must not exist.", c.OutputDir))
		}
	}

	return errs
}

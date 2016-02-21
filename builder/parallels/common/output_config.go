package common

import (
	"fmt"
	"os"
	"path"

	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/template/interpolate"
)

type OutputConfig struct {
	OutputDir string `mapstructure:"output_directory"`
}

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

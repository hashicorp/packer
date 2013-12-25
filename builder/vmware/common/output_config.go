package common

import (
	"fmt"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
	"os"
)

type OutputConfig struct {
	OutputDir string `mapstructure:"output_directory"`
}

func (c *OutputConfig) Prepare(t *packer.ConfigTemplate, pc *common.PackerConfig) []error {
	if c.OutputDir == "" {
		c.OutputDir = fmt.Sprintf("output-%s", pc.PackerBuildName)
	}

	templates := map[string]*string{
		"output_directory": &c.OutputDir,
	}

	errs := make([]error, 0)
	for n, ptr := range templates {
		var err error
		*ptr, err = t.Process(*ptr, nil)
		if err != nil {
			errs = append(errs, fmt.Errorf("Error processing %s: %s", n, err))
		}
	}

	if !pc.PackerForce {
		if _, err := os.Stat(c.OutputDir); err == nil {
			errs = append(errs, fmt.Errorf(
				"Output directory '%s' already exists. It must not exist.", c.OutputDir))
		}
	}

	return errs
}

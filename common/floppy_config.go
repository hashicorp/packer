package common

import (
	"fmt"
	"os"

	"github.com/mitchellh/packer/template/interpolate"
)

type FloppyConfig struct {
	FloppyFiles []string `mapstructure:"floppy_files"`
}

func (c *FloppyConfig) Prepare(ctx *interpolate.Context) []error {
	var errs []error

	if c.FloppyFiles == nil {
		c.FloppyFiles = make([]string, 0)
	}

	for _, path := range c.FloppyFiles {
		if _, err := os.Stat(path); err != nil {
			errs = append(errs, fmt.Errorf("Bad Floppy disk file '%s': %s", path, err))
		}
	}

	return errs
}

package common

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/packer/template/interpolate"
)

type FloppyConfig struct {
	FloppyFiles       []string `mapstructure:"floppy_files"`
	FloppyDirectories []string `mapstructure:"floppy_dirs"`
}

func (c *FloppyConfig) Prepare(ctx *interpolate.Context) []error {
	var errs []error
	var err error

	if c.FloppyFiles == nil {
		c.FloppyFiles = make([]string, 0)
	}

	for _, path := range c.FloppyFiles {
		if strings.IndexAny(path, "*?[") >= 0 {
			_, err = filepath.Glob(path)
		} else {
			_, err = os.Stat(path)
		}
		if err != nil {
			errs = append(errs, fmt.Errorf("Bad Floppy disk file '%s': %s", path, err))
		}
	}

	if c.FloppyDirectories == nil {
		c.FloppyDirectories = make([]string, 0)
	}

	for _, path := range c.FloppyDirectories {
		if strings.IndexAny(path, "*?[") >= 0 {
			_, err = filepath.Glob(path)
		} else {
			_, err = os.Stat(path)
		}
		if err != nil {
			errs = append(errs, fmt.Errorf("Bad Floppy disk directory '%s': %s", path, err))
		}
	}

	return errs
}

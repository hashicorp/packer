package common

import (
	"fmt"

	"github.com/mitchellh/packer/packer"
)

// FloppyConfig is configuration related to created floppy disks and attaching
// them to a VirtualBox machine.
type FloppyConfig struct {
	FloppyFiles []string `mapstructure:"floppy_files"`
}

func (c *FloppyConfig) Prepare(t *packer.ConfigTemplate) []error {
	if c.FloppyFiles == nil {
		c.FloppyFiles = make([]string, 0)
	}

	errs := make([]error, 0)
	for i, file := range c.FloppyFiles {
		var err error
		c.FloppyFiles[i], err = t.Process(file, nil)
		if err != nil {
			errs = append(errs, fmt.Errorf(
				"Error processing floppy_files[%d]: %s", i, err))
		}
	}

	return errs
}

package common

import (
	"fmt"

	"github.com/hashicorp/packer/template/interpolate"
)

type HWConfig struct {

	// cpu information
	CpuCount   int `mapstructure:"cpus"`
	MemorySize int `mapstructure:"memory"`

	// device presence
	Sound bool `mapstructure:"sound"`
	USB   bool `mapstructure:"usb"`
}

func (c *HWConfig) Prepare(ctx *interpolate.Context) []error {
	var errs []error

	// Hardware and cpu options
	if c.CpuCount < 0 {
		errs = append(errs, fmt.Errorf("An invalid number of cpus was specified (cpus < 0): %d", c.CpuCount))
		c.CpuCount = 0
	}
	if c.CpuCount == 0 {
		c.CpuCount = 1
	}

	if c.MemorySize < 0 {
		errs = append(errs, fmt.Errorf("An invalid memory size was specified (memory < 0): %d", c.MemorySize))
		c.MemorySize = 0
	}
	if c.MemorySize == 0 {
		c.MemorySize = 512
	}

	// Peripherals
	if !c.Sound {
		c.Sound = false
	}

	if !c.USB {
		c.USB = false
	}

	return nil
}

package common

import (
	"fmt"
	"time"

	"github.com/mitchellh/packer/packer"
)

type RunConfig struct {
	Headless    bool   `mapstructure:"headless"`
	RawBootWait string `mapstructure:"boot_wait"`

	BootWait time.Duration ``
}

func (c *RunConfig) Prepare(t *packer.ConfigTemplate) []error {
	if c.RawBootWait == "" {
		c.RawBootWait = "10s"
	}

	templates := map[string]*string{
		"boot_wait": &c.RawBootWait,
	}

	var err error
	errs := make([]error, 0)
	for n, ptr := range templates {
		*ptr, err = t.Process(*ptr, nil)
		if err != nil {
			errs = append(errs, fmt.Errorf("Error processing %s: %s", n, err))
		}
	}

	if c.RawBootWait != "" {
		c.BootWait, err = time.ParseDuration(c.RawBootWait)
		if err != nil {
			errs = append(
				errs, fmt.Errorf("Failed parsing boot_wait: %s", err))
		}
	}

	return errs
}

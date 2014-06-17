package common

import (
	"fmt"
	"github.com/mitchellh/packer/packer"
	"time"
)

type ShutdownConfig struct {
	ShutdownCommand    string `mapstructure:"shutdown_command"`
	RawShutdownTimeout string `mapstructure:"shutdown_timeout"`

	ShutdownTimeout time.Duration ``
}

func (c *ShutdownConfig) Prepare(t *packer.ConfigTemplate) []error {
	if c.RawShutdownTimeout == "" {
		c.RawShutdownTimeout = "5m"
	}

	templates := map[string]*string{
		"shutdown_command": &c.ShutdownCommand,
		"shutdown_timeout": &c.RawShutdownTimeout,
	}

	errs := make([]error, 0)
	for n, ptr := range templates {
		var err error
		*ptr, err = t.Process(*ptr, nil)
		if err != nil {
			errs = append(errs, fmt.Errorf("Error processing %s: %s", n, err))
		}
	}

	var err error
	c.ShutdownTimeout, err = time.ParseDuration(c.RawShutdownTimeout)
	if err != nil {
		errs = append(errs, fmt.Errorf("Failed parsing shutdown_timeout: %s", err))
	}

	return errs
}

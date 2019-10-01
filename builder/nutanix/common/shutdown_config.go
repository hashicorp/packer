package common

import (
	"fmt"
	"time"
)

// ShutdownConfig contains the configuration for the shutdown parameters
type ShutdownConfig struct {
	Command    string `mapstructure:"shutdown_command"`
	RawTimeout string `mapstructure:"shutdown_timeout"`

	Timeout time.Duration
}

// Prepare will validate the shutdown config for the image
func (c *ShutdownConfig) Prepare() []error {
	var errs []error

	if c.RawTimeout != "" {
		timeout, err := time.ParseDuration(c.RawTimeout)
		if err != nil {
			errs = append(errs, fmt.Errorf("Failed parsing shutdown_timeout: %s", err))
			return errs
		}
		c.Timeout = timeout
	} else {
		c.Timeout = 5 * time.Minute
	}

	return nil
}

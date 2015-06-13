package communicator

import (
	"errors"
	"time"

	"github.com/mitchellh/packer/template/interpolate"
)

// Config is the common configuration that communicators allow within
// a builder.
type Config struct {
	Type          string        `mapstructure:"communicator"`
	SSHHost       string        `mapstructure:"ssh_host"`
	SSHPort       int           `mapstructure:"ssh_port"`
	SSHUsername   string        `mapstructure:"ssh_username"`
	SSHPassword   string        `mapstructure:"ssh_password"`
	SSHPrivateKey string        `mapstructure:"ssh_private_key_file"`
	SSHPty        bool          `mapstructure:"ssh_pty"`
	SSHTimeout    time.Duration `mapstructure:"ssh_timeout"`
}

func (c *Config) Prepare(ctx *interpolate.Context) []error {
	if c.Type == "" {
		c.Type = "ssh"
	}

	if c.SSHPort == 0 {
		c.SSHPort = 22
	}

	if c.SSHTimeout == 0 {
		c.SSHTimeout = 5 * time.Minute
	}

	// Validation
	var errs []error
	if c.SSHUsername == "" {
		errs = append(errs, errors.New("An ssh_username must be specified"))
	}

	return errs
}

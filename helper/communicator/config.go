package communicator

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/mitchellh/packer/template/interpolate"
)

// Config is the common configuration that communicators allow within
// a builder.
type Config struct {
	Type                 string        `mapstructure:"communicator"`
	SSHHost              string        `mapstructure:"ssh_host"`
	SSHPort              int           `mapstructure:"ssh_port"`
	SSHUsername          string        `mapstructure:"ssh_username"`
	SSHPassword          string        `mapstructure:"ssh_password"`
	SSHPrivateKey        string        `mapstructure:"ssh_private_key_file"`
	SSHPty               bool          `mapstructure:"ssh_pty"`
	SSHTimeout           time.Duration `mapstructure:"ssh_timeout"`
	SSHHandshakeAttempts int           `mapstructure:"ssh_handshake_attempts"`
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

	if c.SSHHandshakeAttempts == 0 {
		c.SSHHandshakeAttempts = 10
	}

	// Validation
	var errs []error
	if c.Type == "ssh" {
		if c.SSHUsername == "" {
			errs = append(errs, errors.New("An ssh_username must be specified"))
		}

		if c.SSHPrivateKey != "" {
			if _, err := os.Stat(c.SSHPrivateKey); err != nil {
				errs = append(errs, fmt.Errorf(
					"ssh_private_key_file is invalid: %s", err))
			} else if _, err := SSHFileSigner(c.SSHPrivateKey); err != nil {
				errs = append(errs, fmt.Errorf(
					"ssh_private_key_file is invalid: %s", err))
			}
		}
	}

	return errs
}

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
	Type string `mapstructure:"communicator"`

	// SSH
	SSHHost       string        `mapstructure:"ssh_host"`
	SSHPort       int           `mapstructure:"ssh_port"`
	SSHUsername   string        `mapstructure:"ssh_username"`
	SSHPassword   string        `mapstructure:"ssh_password"`
	SSHPrivateKey string        `mapstructure:"ssh_private_key_file"`
	SSHPty        bool          `mapstructure:"ssh_pty"`
	SSHTimeout    time.Duration `mapstructure:"ssh_timeout"`

	// WinRM
	WinRMUser     string        `mapstructure:"winrm_username"`
	WinRMPassword string        `mapstructure:"winrm_password"`
	WinRMHost     string        `mapstructure:"winrm_host"`
	WinRMPort     int           `mapstructure:"winrm_port"`
	WinRMTimeout  time.Duration `mapstructure:"winrm_timeout"`
}

// Port returns the port that will be used for access based on config.
func (c *Config) Port() int {
	switch c.Type {
	case "ssh":
		return c.SSHPort
	case "winrm":
		return c.WinRMPort
	default:
		return 0
	}
}

func (c *Config) Prepare(ctx *interpolate.Context) []error {
	if c.Type == "" {
		c.Type = "ssh"
	}

	var errs []error
	switch c.Type {
	case "ssh":
		if es := c.prepareSSH(ctx); len(es) > 0 {
			errs = append(errs, es...)
		}
	case "winrm":
		if es := c.prepareWinRM(ctx); len(es) > 0 {
			errs = append(errs, es...)
		}
	}

	return errs
}

func (c *Config) prepareSSH(ctx *interpolate.Context) []error {
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

	if c.SSHPrivateKey != "" {
		if _, err := os.Stat(c.SSHPrivateKey); err != nil {
			errs = append(errs, fmt.Errorf(
				"ssh_private_key_file is invalid: %s", err))
		} else if _, err := SSHFileSigner(c.SSHPrivateKey); err != nil {
			errs = append(errs, fmt.Errorf(
				"ssh_private_key_file is invalid: %s", err))
		}
	}

	return errs
}

func (c *Config) prepareWinRM(ctx *interpolate.Context) []error {
	if c.WinRMPort == 0 {
		c.WinRMPort = 5985
	}

	if c.WinRMTimeout == 0 {
		c.WinRMTimeout = 30 * time.Minute
	}

	var errs []error
	if c.WinRMUser == "" {
		errs = append(errs, errors.New("winrm_username must be specified."))
	}

	return errs
}

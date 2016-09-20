package communicator

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/mitchellh/packer/template/interpolate"
)

// Config is the common configuration that communicators allow within
// a builder.
type Config struct {
	Type string `mapstructure:"communicator"`

	// SSH
	SSHHost               string        `mapstructure:"ssh_host"`
	SSHPort               int           `mapstructure:"ssh_port"`
	SSHUsername           string        `mapstructure:"ssh_username"`
	SSHPassword           string        `mapstructure:"ssh_password"`
	SSHPrivateKey         string        `mapstructure:"ssh_private_key_file"`
	SSHPty                bool          `mapstructure:"ssh_pty"`
	SSHTimeout            time.Duration `mapstructure:"ssh_timeout"`
	SSHDisableAgent       bool          `mapstructure:"ssh_disable_agent"`
	SSHHandshakeAttempts  int           `mapstructure:"ssh_handshake_attempts"`
	SSHBastionHost        string        `mapstructure:"ssh_bastion_host"`
	SSHBastionPort        int           `mapstructure:"ssh_bastion_port"`
	SSHBastionUsername    string        `mapstructure:"ssh_bastion_username"`
	SSHBastionPassword    string        `mapstructure:"ssh_bastion_password"`
	SSHBastionPrivateKey  string        `mapstructure:"ssh_bastion_private_key_file"`
	SSHFileTransferMethod string        `mapstructure:"ssh_file_transfer_method"`

	// WinRM
	WinRMUser               string        `mapstructure:"winrm_username"`
	WinRMPassword           string        `mapstructure:"winrm_password"`
	WinRMHost               string        `mapstructure:"winrm_host"`
	WinRMPort               int           `mapstructure:"winrm_port"`
	WinRMTimeout            time.Duration `mapstructure:"winrm_timeout"`
	WinRMUseSSL             bool          `mapstructure:"winrm_use_ssl"`
	WinRMInsecure           bool          `mapstructure:"winrm_insecure"`
	WinRMTransportDecorator func(*http.Transport) http.RoundTripper
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

// Host returns the port that will be used for access based on config.
func (c *Config) Host() string {
	switch c.Type {
	case "ssh":
		return c.SSHHost
	case "winrm":
		return c.WinRMHost
	default:
		return ""
	}
}

// User returns the port that will be used for access based on config.
func (c *Config) User() string {
	switch c.Type {
	case "ssh":
		return c.SSHUsername
	case "winrm":
		return c.WinRMUser
	default:
		return ""
	}
}

// Password returns the port that will be used for access based on config.
func (c *Config) Password() string {
	switch c.Type {
	case "ssh":
		return c.SSHPassword
	case "winrm":
		return c.WinRMPassword
	default:
		return ""
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
	case "docker", "none":
		break
	default:
		return []error{fmt.Errorf("Communicator type %s is invalid", c.Type)}
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

	if c.SSHHandshakeAttempts == 0 {
		c.SSHHandshakeAttempts = 10
	}

	if c.SSHBastionHost != "" {
		if c.SSHBastionPort == 0 {
			c.SSHBastionPort = 22
		}

		if c.SSHBastionPrivateKey == "" && c.SSHPrivateKey != "" {
			c.SSHBastionPrivateKey = c.SSHPrivateKey
		}
	}

	if c.SSHFileTransferMethod == "" {
		c.SSHFileTransferMethod = "scp"
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

	if c.SSHBastionHost != "" {
		if c.SSHBastionPassword == "" && c.SSHBastionPrivateKey == "" {
			errs = append(errs, errors.New(
				"ssh_bastion_password or ssh_bastion_private_key_file must be specified"))
		}
	}

	if c.SSHFileTransferMethod != "scp" && c.SSHFileTransferMethod != "sftp" {
		errs = append(errs, fmt.Errorf(
			"ssh_file_transfer_method ('%s') is invalid, valid methods: sftp, scp",
			c.SSHFileTransferMethod))
	}

	return errs
}

func (c *Config) prepareWinRM(ctx *interpolate.Context) []error {
	if c.WinRMPort == 0 && c.WinRMUseSSL {
		c.WinRMPort = 5986
	} else if c.WinRMPort == 0 {
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

package communicator

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/hashicorp/packer/template/interpolate"
	"github.com/masterzen/winrm"
)

// Config is the common configuration that communicators allow within
// a builder.
type Config struct {
	Type string `mapstructure:"communicator"`

	// SSH
	SSHHost                   string        `mapstructure:"ssh_host"`
	SSHPort                   int           `mapstructure:"ssh_port"`
	SSHUsername               string        `mapstructure:"ssh_username"`
	SSHPassword               string        `mapstructure:"ssh_password"`
	SSHPrivateKey             string        `mapstructure:"ssh_private_key_file"`
	SSHPty                    bool          `mapstructure:"ssh_pty"`
	SSHTimeout                time.Duration `mapstructure:"ssh_timeout"`
	SSHAgentAuth              bool          `mapstructure:"ssh_agent_auth"`
	SSHDisableAgentForwarding bool          `mapstructure:"ssh_disable_agent_forwarding"`
	SSHHandshakeAttempts      int           `mapstructure:"ssh_handshake_attempts"`
	SSHBastionHost            string        `mapstructure:"ssh_bastion_host"`
	SSHBastionPort            int           `mapstructure:"ssh_bastion_port"`
	SSHBastionAgentAuth       bool          `mapstructure:"ssh_bastion_agent_auth"`
	SSHBastionUsername        string        `mapstructure:"ssh_bastion_username"`
	SSHBastionPassword        string        `mapstructure:"ssh_bastion_password"`
	SSHBastionPrivateKey      string        `mapstructure:"ssh_bastion_private_key_file"`
	SSHFileTransferMethod     string        `mapstructure:"ssh_file_transfer_method"`
	SSHProxyHost              string        `mapstructure:"ssh_proxy_host"`
	SSHProxyPort              int           `mapstructure:"ssh_proxy_port"`
	SSHProxyUsername          string        `mapstructure:"ssh_proxy_username"`
	SSHProxyPassword          string        `mapstructure:"ssh_proxy_password"`
	SSHKeepAliveInterval      time.Duration `mapstructure:"ssh_keep_alive_interval"`
	SSHReadWriteTimeout       time.Duration `mapstructure:"ssh_read_write_timeout"`

	// WinRM
	WinRMUser               string        `mapstructure:"winrm_username"`
	WinRMPassword           string        `mapstructure:"winrm_password"`
	WinRMHost               string        `mapstructure:"winrm_host"`
	WinRMPort               int           `mapstructure:"winrm_port"`
	WinRMTimeout            time.Duration `mapstructure:"winrm_timeout"`
	WinRMUseSSL             bool          `mapstructure:"winrm_use_ssl"`
	WinRMInsecure           bool          `mapstructure:"winrm_insecure"`
	WinRMUseNTLM            bool          `mapstructure:"winrm_use_ntlm"`
	WinRMTransportDecorator func() winrm.Transporter
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

	if c.SSHKeepAliveInterval == 0 {
		c.SSHKeepAliveInterval = 5 * time.Second
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

	if c.SSHProxyHost != "" {
		if c.SSHProxyPort == 0 {
			c.SSHProxyPort = 1080
		}
	}

	if c.SSHFileTransferMethod == "" {
		c.SSHFileTransferMethod = "scp"
	}

	// Validation
	var errs []error
	if c.SSHUsername == "" {
		errs = append(errs, errors.New("An ssh_username must be specified\n  Note: some builders used to default ssh_username to \"root\"."))
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

	if c.SSHBastionHost != "" && !c.SSHBastionAgentAuth {
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

	if c.SSHBastionHost != "" && c.SSHProxyHost != "" {
		errs = append(errs, errors.New("please specify either ssh_bastion_host or ssh_proxy_host, not both"))
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

	if c.WinRMUseNTLM == true {
		c.WinRMTransportDecorator = func() winrm.Transporter { return &winrm.ClientNTLM{} }
	}

	var errs []error
	if c.WinRMUser == "" {
		errs = append(errs, errors.New("winrm_username must be specified."))
	}

	return errs
}

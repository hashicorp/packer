package common

import (
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	commonssh "github.com/mitchellh/packer/common/ssh"
	"github.com/mitchellh/packer/packer"
)

type SSHConfig struct {
	SSHUser           string `mapstructure:"ssh_username"`
	SSHKeyPath        string `mapstructure:"ssh_key_path"`
	SSHPassword       string `mapstructure:"ssh_password"`
	SSHHost           string `mapstructure:"ssh_host"`
	SSHPort           uint   `mapstructure:"ssh_port"`
	SSHSkipRequestPty bool   `mapstructure:"ssh_skip_request_pty"`
	RawSSHWaitTimeout string `mapstructure:"ssh_wait_timeout"`

	SSHWaitTimeout time.Duration
}

func (c *SSHConfig) Prepare(t *packer.ConfigTemplate) []error {
	if c.SSHPort == 0 {
		c.SSHPort = 22
	}

	if c.RawSSHWaitTimeout == "" {
		c.RawSSHWaitTimeout = "20m"
	}

	templates := map[string]*string{
		"ssh_host":         &c.SSHHost,
		"ssh_key_path":     &c.SSHKeyPath,
		"ssh_password":     &c.SSHPassword,
		"ssh_username":     &c.SSHUser,
		"ssh_wait_timeout": &c.RawSSHWaitTimeout,
	}

	errs := make([]error, 0)
	for n, ptr := range templates {
		var err error
		*ptr, err = t.Process(*ptr, nil)
		if err != nil {
			errs = append(errs, fmt.Errorf("Error processing %s: %s", n, err))
		}
	}

	if c.SSHKeyPath != "" {
		if _, err := os.Stat(c.SSHKeyPath); err != nil {
			errs = append(errs, fmt.Errorf("ssh_key_path is invalid: %s", err))
		} else if _, err := commonssh.FileSigner(c.SSHKeyPath); err != nil {
			errs = append(errs, fmt.Errorf("ssh_key_path is invalid: %s", err))
		}
	}

	if c.SSHHost != "" {
		if ip := net.ParseIP(c.SSHHost); ip == nil {
			if _, err := net.LookupHost(c.SSHHost); err != nil {
				errs = append(errs, errors.New("ssh_host is an invalid IP or hostname"))
			}
		}
	}

	if c.SSHUser == "" {
		errs = append(errs, errors.New("An ssh_username must be specified."))
	}

	var err error
	c.SSHWaitTimeout, err = time.ParseDuration(c.RawSSHWaitTimeout)
	if err != nil {
		errs = append(errs, fmt.Errorf("Failed parsing ssh_wait_timeout: %s", err))
	}

	return errs
}

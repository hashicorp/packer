package common

import (
	"errors"
	"fmt"
	"os"
	"time"

	commonssh "github.com/mitchellh/packer/common/ssh"
	"github.com/mitchellh/packer/template/interpolate"
)

type SSHConfig struct {
	SSHHostPortMin    uint   `mapstructure:"ssh_host_port_min"`
	SSHHostPortMax    uint   `mapstructure:"ssh_host_port_max"`
	SSHKeyPath        string `mapstructure:"ssh_key_path"`
	SSHPassword       string `mapstructure:"ssh_password"`
	SSHPort           uint   `mapstructure:"ssh_port"`
	SSHUser           string `mapstructure:"ssh_username"`
	RawSSHWaitTimeout string `mapstructure:"ssh_wait_timeout"`
	SSHSkipNatMapping bool   `mapstructure:"ssh_skip_nat_mapping"`

	SSHWaitTimeout time.Duration
}

func (c *SSHConfig) Prepare(ctx *interpolate.Context) []error {
	if c.SSHHostPortMin == 0 {
		c.SSHHostPortMin = 2222
	}

	if c.SSHHostPortMax == 0 {
		c.SSHHostPortMax = 4444
	}

	if c.SSHPort == 0 {
		c.SSHPort = 22
	}

	if c.RawSSHWaitTimeout == "" {
		c.RawSSHWaitTimeout = "20m"
	}

	var errs []error
	if c.SSHKeyPath != "" {
		if _, err := os.Stat(c.SSHKeyPath); err != nil {
			errs = append(errs, fmt.Errorf("ssh_key_path is invalid: %s", err))
		} else if _, err := commonssh.FileSigner(c.SSHKeyPath); err != nil {
			errs = append(errs, fmt.Errorf("ssh_key_path is invalid: %s", err))
		}
	}

	if c.SSHHostPortMin > c.SSHHostPortMax {
		errs = append(errs,
			errors.New("ssh_host_port_min must be less than ssh_host_port_max"))
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

package common

import (
	"errors"
	"fmt"
	"github.com/mitchellh/goamz/aws"
	"time"
)

// RunConfig contains configuration for running an instance from a source
// AMI and details on how to access that launched image.
type RunConfig struct {
	Region          string
	SourceAmi       string `mapstructure:"source_ami"`
	InstanceType    string `mapstructure:"instance_type"`
	RawSSHTimeout   string `mapstructure:"ssh_timeout"`
	SSHUsername     string `mapstructure:"ssh_username"`
	SSHPort         int    `mapstructure:"ssh_port"`
	SecurityGroupId string `mapstructure:"security_group_id"`
	SubnetId        string `mapstructure:"subnet_id"`
	VpcId           string `mapstructure:"vpc_id"`

	// Unexported fields that are calculated from others
	sshTimeout time.Duration
}

func (c *RunConfig) Prepare() []error {
	// Defaults
	if c.SSHPort == 0 {
		c.SSHPort = 22
	}

	if c.RawSSHTimeout == "" {
		c.RawSSHTimeout = "1m"
	}

	// Validation
	var err error
	errs := make([]error, 0)
	if c.SourceAmi == "" {
		errs = append(errs, errors.New("A source_ami must be specified"))
	}

	if c.InstanceType == "" {
		errs = append(errs, errors.New("An instance_type must be specified"))
	}

	if c.Region == "" {
		errs = append(errs, errors.New("A region must be specified"))
	} else if _, ok := aws.Regions[c.Region]; !ok {
		errs = append(errs, fmt.Errorf("Unknown region: %s", c.Region))
	}

	if c.SSHUsername == "" {
		errs = append(errs, errors.New("An ssh_username must be specified"))
	}

	c.sshTimeout, err = time.ParseDuration(c.RawSSHTimeout)
	if err != nil {
		errs = append(errs, fmt.Errorf("Failed parsing ssh_timeout: %s", err))
	}

	return errs
}

func (c *RunConfig) SSHTimeout() time.Duration {
	return c.sshTimeout
}

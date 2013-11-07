package common

import (
	"errors"
	"fmt"
	"github.com/mitchellh/packer/packer"
	"os"
	"time"
)

// RunConfig contains configuration for running an instance from a source
// AMI and details on how to access that launched image.
type RunConfig struct {
	SourceAmi            string   `mapstructure:"source_ami"`
	IamInstanceProfile   string   `mapstructure:"iam_instance_profile"`
	InstanceType         string   `mapstructure:"instance_type"`
	UserData             string   `mapstructure:"user_data"`
	UserDataFile         string   `mapstructure:"user_data_file"`
	RawSSHTimeout        string   `mapstructure:"ssh_timeout"`
	SSHUsername          string   `mapstructure:"ssh_username"`
	SSHPort              int      `mapstructure:"ssh_port"`
	SecurityGroupId      string   `mapstructure:"security_group_id"`
	SecurityGroupIds     []string `mapstructure:"security_group_ids"`
	SubnetId             string   `mapstructure:"subnet_id"`
	TemporaryKeyPairName string   `mapstructure:"temporary_key_pair_name"`
	VpcId                string   `mapstructure:"vpc_id"`
	AvailabilityZone     string   `mapstructure:"availability_zone"`

	// Unexported fields that are calculated from others
	sshTimeout time.Duration
}

func (c *RunConfig) Prepare(t *packer.ConfigTemplate) []error {
	if t == nil {
		var err error
		t, err = packer.NewConfigTemplate()
		if err != nil {
			return []error{err}
		}
	}

	// Defaults
	if c.SSHPort == 0 {
		c.SSHPort = 22
	}

	if c.RawSSHTimeout == "" {
		c.RawSSHTimeout = "1m"
	}

	if c.TemporaryKeyPairName == "" {
		c.TemporaryKeyPairName = "packer {{uuid}}"
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

	if c.SSHUsername == "" {
		errs = append(errs, errors.New("An ssh_username must be specified"))
	}

	if c.UserData != "" && c.UserDataFile != "" {
		errs = append(errs, fmt.Errorf("Only one of user_data or user_data_file can be specified."))
	} else if c.UserDataFile != "" {
		if _, err := os.Stat(c.UserDataFile); err != nil {
			errs = append(errs, fmt.Errorf("user_data_file not found: %s", c.UserDataFile))
		}
	}

	if c.SecurityGroupId != "" {
		if len(c.SecurityGroupIds) > 0 {
			errs = append(errs, fmt.Errorf("Only one of security_group_id or security_group_ids can be specified."))
		} else {
			c.SecurityGroupIds = []string{c.SecurityGroupId}
			c.SecurityGroupId = ""
		}
	}

	templates := map[string]*string{
		"iam_instance_profile":    &c.IamInstanceProfile,
		"instance_type":           &c.InstanceType,
		"ssh_timeout":             &c.RawSSHTimeout,
		"ssh_username":            &c.SSHUsername,
		"source_ami":              &c.SourceAmi,
		"subnet_id":               &c.SubnetId,
		"temporary_key_pair_name": &c.TemporaryKeyPairName,
		"vpc_id":                  &c.VpcId,
		"availability_zone":       &c.AvailabilityZone,
	}

	for n, ptr := range templates {
		var err error
		*ptr, err = t.Process(*ptr, nil)
		if err != nil {
			errs = append(
				errs, fmt.Errorf("Error processing %s: %s", n, err))
		}
	}

	sliceTemplates := map[string][]string{
		"security_group_ids": c.SecurityGroupIds,
	}

	for n, slice := range sliceTemplates {
		for i, elem := range slice {
			var err error
			slice[i], err = t.Process(elem, nil)
			if err != nil {
				errs = append(
					errs, fmt.Errorf("Error processing %s[%d]: %s", n, i, err))
			}
		}
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

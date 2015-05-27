package common

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/mitchellh/packer/common/uuid"
	"github.com/mitchellh/packer/template/interpolate"
)

// RunConfig contains configuration for running an instance from a source
// AMI and details on how to access that launched image.
type RunConfig struct {
	AssociatePublicIpAddress bool              `mapstructure:"associate_public_ip_address"`
	AvailabilityZone         string            `mapstructure:"availability_zone"`
	IamInstanceProfile       string            `mapstructure:"iam_instance_profile"`
	InstanceType             string            `mapstructure:"instance_type"`
	RunTags                  map[string]string `mapstructure:"run_tags"`
	SourceAmi                string            `mapstructure:"source_ami"`
	SpotPrice                string            `mapstructure:"spot_price"`
	SpotPriceAutoProduct     string            `mapstructure:"spot_price_auto_product"`
	RawSSHTimeout            string            `mapstructure:"ssh_timeout"`
	SSHUsername              string            `mapstructure:"ssh_username"`
	SSHPrivateKeyFile        string            `mapstructure:"ssh_private_key_file"`
	SSHPrivateIp             bool              `mapstructure:"ssh_private_ip"`
	SSHPort                  int               `mapstructure:"ssh_port"`
	SecurityGroupId          string            `mapstructure:"security_group_id"`
	SecurityGroupIds         []string          `mapstructure:"security_group_ids"`
	SubnetId                 string            `mapstructure:"subnet_id"`
	TemporaryKeyPairName     string            `mapstructure:"temporary_key_pair_name"`
	UserData                 string            `mapstructure:"user_data"`
	UserDataFile             string            `mapstructure:"user_data_file"`
	VpcId                    string            `mapstructure:"vpc_id"`

	// Unexported fields that are calculated from others
	sshTimeout time.Duration
}

func (c *RunConfig) Prepare(ctx *interpolate.Context) []error {
	// Defaults
	if c.SSHPort == 0 {
		c.SSHPort = 22
	}

	if c.RawSSHTimeout == "" {
		c.RawSSHTimeout = "5m"
	}

	if c.TemporaryKeyPairName == "" {
		c.TemporaryKeyPairName = fmt.Sprintf(
			"packer %s", uuid.TimeOrderedUUID())
	}

	// Validation
	var errs []error
	if c.SourceAmi == "" {
		errs = append(errs, errors.New("A source_ami must be specified"))
	}

	if c.InstanceType == "" {
		errs = append(errs, errors.New("An instance_type must be specified"))
	}

	if c.SpotPrice == "auto" {
		if c.SpotPriceAutoProduct == "" {
			errs = append(errs, errors.New(
				"spot_price_auto_product must be specified when spot_price is auto"))
		}
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

	var err error
	c.sshTimeout, err = time.ParseDuration(c.RawSSHTimeout)
	if err != nil {
		errs = append(errs, fmt.Errorf("Failed parsing ssh_timeout: %s", err))
	}

	return errs
}

func (c *RunConfig) SSHTimeout() time.Duration {
	return c.sshTimeout
}

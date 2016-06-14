package common

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/mitchellh/packer/common/uuid"
	"github.com/mitchellh/packer/helper/communicator"
	"github.com/mitchellh/packer/template/interpolate"
)

var reShutdownBehavior = regexp.MustCompile("^(stop|terminate)$")

// RunConfig contains configuration for running an instance from a source
// AMI and details on how to access that launched image.
type RunConfig struct {
	AssociatePublicIpAddress          bool              `mapstructure:"associate_public_ip_address"`
	AvailabilityZone                  string            `mapstructure:"availability_zone"`
	EbsOptimized                      bool              `mapstructure:"ebs_optimized"`
	IamInstanceProfile                string            `mapstructure:"iam_instance_profile"`
	InstanceType                      string            `mapstructure:"instance_type"`
	RunTags                           map[string]string `mapstructure:"run_tags"`
	SourceAmi                         string            `mapstructure:"source_ami"`
	SpotPrice                         string            `mapstructure:"spot_price"`
	SpotPriceAutoProduct              string            `mapstructure:"spot_price_auto_product"`
	DisableStopInstance               bool              `mapstructure:"disable_stop_instance"`
	SecurityGroupId                   string            `mapstructure:"security_group_id"`
	SecurityGroupIds                  []string          `mapstructure:"security_group_ids"`
	SubnetId                          string            `mapstructure:"subnet_id"`
	TemporaryKeyPairName              string            `mapstructure:"temporary_key_pair_name"`
	UserData                          string            `mapstructure:"user_data"`
	UserDataFile                      string            `mapstructure:"user_data_file"`
	WindowsPasswordTimeout            time.Duration     `mapstructure:"windows_password_timeout"`
	VpcId                             string            `mapstructure:"vpc_id"`
	InstanceInitiatedShutdownBehavior string            `mapstructure:"shutdown_behaviour"`

	// Communicator settings
	Comm           communicator.Config `mapstructure:",squash"`
	SSHKeyPairName string              `mapstructure:"ssh_keypair_name"`
	SSHPrivateIp   bool                `mapstructure:"ssh_private_ip"`
}

func (c *RunConfig) Prepare(ctx *interpolate.Context) []error {
	// if we are not given an explicit keypairname, create a temporary one
	if c.SSHKeyPairName == "" {
		c.TemporaryKeyPairName = fmt.Sprintf(
			"packer %s", uuid.TimeOrderedUUID())
	}

	if c.WindowsPasswordTimeout == 0 {
		c.WindowsPasswordTimeout = 10 * time.Minute
	}

	// Validation
	errs := c.Comm.Prepare(ctx)
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

	if c.InstanceInitiatedShutdownBehavior == "" {
		c.InstanceInitiatedShutdownBehavior = "stop"
	} else if !reShutdownBehavior.MatchString(c.InstanceInitiatedShutdownBehavior) {
		errs = append(errs, fmt.Errorf("shutdown_behaviour only accepts 'stop' or 'terminate' values."))
	}

	return errs
}

package ecs

import (
	"errors"
	"fmt"
	"github.com/mitchellh/packer/common/uuid"
	"github.com/mitchellh/packer/helper/communicator"
	"github.com/mitchellh/packer/template/interpolate"
	"os"
)

const (
	ssh_time_out      = 60000000000
	default_port      = 22
	default_comm_type = "ssh"
)

type RunConfig struct {
	AssociatePublicIpAddress bool   `mapstructure:"associate_public_ip_address"`
	ZoneId                   string `mapstructure:"zone_id"`
	IOOptimized              bool   `mapstructure:"io_optimized"`
	InstanceType             string `mapstructure:"instance_type"`
	Description              string `mapstructure:"description"`
	AlicloudSourceImage      string `mapstructure:"source_image"`
	ForceStopInstance        bool   `mapstructure:"force_stop_instance"`
	SecurityGroupId          string `mapstructure:"security_group_id"`
	SecurityGroupName        string `mapstructure:"security_group_name"`
	UserData                 string `mapstructure:"user_data"`
	UserDataFile             string `mapstructure:"user_data_file"`
	VpcId                    string `mapstructure:"vpc_id"`
	VpcName                  string `mapstructure:"vpc_name"`
	CidrBlock                string `mapstructure:"vpc_cidr_block"`
	VSwitchId                string `mapstructure:"vswitch_id"`
	VSwitchName              string `mapstructure:"vswitch_id"`
	InstanceName             string `mapstructure:"instance_name"`
	InternetChargeType       string `mapstructure:"internet_charge_type"`
	InternetMaxBandwidthOut  int    `mapstructure:"internet_max_bandwith_out"`
	TemporaryKeyPairName     string `mapstructure:"temporary_key_pair_name"`

	// Communicator settings
	Comm           communicator.Config `mapstructure:",squash"`
	SSHKeyPairName string              `mapstructure:"ssh_keypair_name"`
	SSHPrivateIp   bool                `mapstructure:"ssh_private_ip"`
	PublicKey      string              `mapstructure:"ssh_private_key_file"`
}

func (c *RunConfig) Prepare(ctx *interpolate.Context) []error {
	if c.SSHKeyPairName == "" && c.TemporaryKeyPairName == "" &&
		c.Comm.SSHPrivateKey == "" && c.Comm.SSHPassword == "" {

		c.TemporaryKeyPairName = fmt.Sprintf("packer_%s", uuid.TimeOrderedUUID())
	}

	if c.Comm.Type == "" {
		c.Comm.Type = default_comm_type
	}

	if c.Comm.SSHTimeout == 0 {
		c.Comm.SSHTimeout = ssh_time_out
	}

	if c.Comm.SSHPort == 0 {
		c.Comm.SSHPort = default_port
	}

	// Validation
	errs := c.Comm.Prepare(ctx)
	if c.AlicloudSourceImage == "" {
		errs = append(errs, errors.New("A source_image must be specified"))
	}

	if c.InstanceType == "" {
		errs = append(errs, errors.New("An aliclod_instance_type must be specified"))
	}

	if c.UserData != "" && c.UserDataFile != "" {
		errs = append(errs, fmt.Errorf("Only one of user_data or user_data_file can be specified."))
	} else if c.UserDataFile != "" {
		if _, err := os.Stat(c.UserDataFile); err != nil {
			errs = append(errs, fmt.Errorf("user_data_file not found: %s", c.UserDataFile))
		}
	}

	return errs
}

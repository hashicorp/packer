package ecs

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/packer/common/uuid"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/template/interpolate"
)

type RunConfig struct {
	AssociatePublicIpAddress bool   `mapstructure:"associate_public_ip_address"`
	ZoneId                   string `mapstructure:"zone_id"`
	IOOptimized              bool   `mapstructure:"io_optimized"`
	InstanceType             string `mapstructure:"instance_type"`
	Description              string `mapstructure:"description"`
	AlicloudSourceImage      string `mapstructure:"source_image"`
	ForceStopInstance        bool   `mapstructure:"force_stop_instance"`
	DisableStopInstance      bool   `mapstructure:"disable_stop_instance"`
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
	InternetMaxBandwidthOut  int    `mapstructure:"internet_max_bandwidth_out"`
	WaitSnapshotReadyTimeout int    `mapstructure:"wait_snapshot_ready_timeout"`

	// Communicator settings
	Comm         communicator.Config `mapstructure:",squash"`
	SSHPrivateIp bool                `mapstructure:"ssh_private_ip"`
}

func (c *RunConfig) Prepare(ctx *interpolate.Context) []error {
	if c.Comm.SSHKeyPairName == "" && c.Comm.SSHTemporaryKeyPairName == "" &&
		c.Comm.SSHPrivateKeyFile == "" && c.Comm.SSHPassword == "" && c.Comm.WinRMPassword == "" {

		c.Comm.SSHTemporaryKeyPairName = fmt.Sprintf("packer_%s", uuid.TimeOrderedUUID())
	}

	// Validation
	errs := c.Comm.Prepare(ctx)
	if c.AlicloudSourceImage == "" {
		errs = append(errs, errors.New("A source_image must be specified"))
	}

	if strings.TrimSpace(c.AlicloudSourceImage) != c.AlicloudSourceImage {
		errs = append(errs, errors.New("The source_image can't include spaces"))
	}

	if c.InstanceType == "" {
		errs = append(errs, errors.New("An alicloud_instance_type must be specified"))
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

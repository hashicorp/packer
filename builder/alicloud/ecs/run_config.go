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
	// ID of the zone to which the disk belongs.
	ZoneId                   string `mapstructure:"zone_id" required:"false"`
	// Whether an ECS instance is I/O optimized or not.
    // The default value is false.
	IOOptimized              bool   `mapstructure:"io_optimized" required:"false"`
	// Type of the instance. For values, see Instance
    // Type
    // Table.
    // You can also obtain the latest instance type table by invoking the
    // Querying Instance Type
    // Table
    // interface.
	InstanceType             string `mapstructure:"instance_type" required:"true"`
	Description              string `mapstructure:"description"`
	// This is the base image id which you want to
    // create your customized images.
	AlicloudSourceImage      string `mapstructure:"source_image" required:"true"`
	// Whether to force shutdown upon device
    // restart. The default value is false.
	ForceStopInstance        bool   `mapstructure:"force_stop_instance" required:"false"`
	// If this option is set to true, Packer
    // will not stop the instance for you, and you need to make sure the instance
    // will be stopped in the final provisioner command. Otherwise, Packer will
    // timeout while waiting the instance to be stopped. This option is provided
    // for some specific scenarios that you want to stop the instance by yourself.
    // E.g., Sysprep a windows which may shutdown the instance within its command.
    // The default value is false.
	DisableStopInstance      bool   `mapstructure:"disable_stop_instance" required:"false"`
	// ID of the security group to which a newly
    // created instance belongs. Mutual access is allowed between instances in one
    // security group. If not specified, the newly created instance will be added
    // to the default security group. If the default group doesn’t exist, or the
    // number of instances in it has reached the maximum limit, a new security
    // group will be created automatically.
	SecurityGroupId          string `mapstructure:"security_group_id" required:"false"`
	// The security group name. The default value
    // is blank. [2, 128] English or Chinese characters, must begin with an
    // uppercase/lowercase letter or Chinese character. Can contain numbers, .,
    // _ or -. It cannot begin with http:// or https://.
	SecurityGroupName        string `mapstructure:"security_group_name" required:"false"`
	// User data to apply when launching the instance. Note
    // that you need to be careful about escaping characters due to the templates
    // being JSON. It is often more convenient to use user_data_file, instead.
    // Packer will not automatically wait for a user script to finish before
    // shutting down the instance this must be handled in a provisioner.
	UserData                 string `mapstructure:"user_data" required:"false"`
	// Path to a file that will be used for the user
    // data when launching the instance.
	UserDataFile             string `mapstructure:"user_data_file" required:"false"`
	// VPC ID allocated by the system.
	VpcId                    string `mapstructure:"vpc_id" required:"false"`
	// The VPC name. The default value is blank. [2, 128]
    // English or Chinese characters, must begin with an uppercase/lowercase
    // letter or Chinese character. Can contain numbers, _ and -. The disk
    // description will appear on the console. Cannot begin with http:// or
    // https://.
	VpcName                  string `mapstructure:"vpc_name" required:"false"`
	// Value options: 192.168.0.0/16 and
    // 172.16.0.0/16. When not specified, the default value is 172.16.0.0/16.
	CidrBlock                string `mapstructure:"vpc_cidr_block" required:"false"`
	// The ID of the VSwitch to be used.
	VSwitchId                string `mapstructure:"vswitch_id" required:"false"`
	// The ID of the VSwitch to be used.
	VSwitchName              string `mapstructure:"vswitch_id" required:"false"`
	// Display name of the instance, which is a string
    // of 2 to 128 Chinese or English characters. It must begin with an
    // uppercase/lowercase letter or a Chinese character and can contain numerals,
    // ., _, or -. The instance name is displayed on the Alibaba Cloud
    // console. If this parameter is not specified, the default value is
    // InstanceId of the instance. It cannot begin with http:// or https://.
	InstanceName             string `mapstructure:"instance_name" required:"false"`
	// Internet charge type, which can be
    // PayByTraffic or PayByBandwidth. Optional values:
	InternetChargeType       string `mapstructure:"internet_charge_type" required:"false"`
	// Maximum outgoing bandwidth to the
    // public network, measured in Mbps (Mega bits per second).
	InternetMaxBandwidthOut  int    `mapstructure:"internet_max_bandwidth_out" required:"false"`
	// Timeout of creating snapshot(s).
    // The default timeout is 3600 seconds if this option is not set or is set
    // to 0. For those disks containing lots of data, it may require a higher
    // timeout value.
	WaitSnapshotReadyTimeout int    `mapstructure:"wait_snapshot_ready_timeout" required:"false"`

	// Communicator settings
	Comm         communicator.Config `mapstructure:",squash"`
	// If this value is true, packer will connect to
    // the ECS created through private ip instead of allocating a public ip or an
    // EIP. The default value is false.
	SSHPrivateIp bool                `mapstructure:"ssh_private_ip" required:"false"`
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

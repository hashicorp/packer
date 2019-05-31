//go:generate struct-markdown

package cvm

import (
	"fmt"
	"os"

	"github.com/hashicorp/packer/common/uuid"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/template/interpolate"
	"github.com/pkg/errors"
)

type TencentCloudRunConfig struct {
	// Whether allocate public ip to your cvm.
    // Default value is false.
	AssociatePublicIpAddress bool   `mapstructure:"associate_public_ip_address" required:"false"`
	// The base image id of Image you want to create
    // your customized image from.
	SourceImageId            string `mapstructure:"source_image_id" required:"true"`
	// The instance type your cvm will be launched by.
    // You should reference Instace Type
    //  for parameter taking.
	InstanceType             string `mapstructure:"instance_type" required:"true"`
	// Instance name.
	InstanceName             string `mapstructure:"instance_name" required:"false"`
	// Root disk type your cvm will be launched by. you could
    // reference Disk Type
    // for parameter taking.
	DiskType                 string `mapstructure:"disk_type" required:"false"`
	// Root disk size your cvm will be launched by. values range(in GB):
	DiskSize                 int64  `mapstructure:"disk_size" required:"false"`
	// Specify vpc your cvm will be launched by.
	VpcId                    string `mapstructure:"vpc_id" required:"false"`
	// Specify vpc name you will create. if vpc_id is not set, packer will
    // create a vpc for you named this parameter.
	VpcName                  string `mapstructure:"vpc_name" required:"false"`
	VpcIp                    string `mapstructure:"vpc_ip"`
	// Specify subnet your cvm will be launched by.
	SubnetId                 string `mapstructure:"subnet_id" required:"false"`
	// Specify subnet name you will create. if subnet_id is not set, packer will
    // create a subnet for you named this parameter.
	SubnetName               string `mapstructure:"subnet_name" required:"false"`
	// Specify cider block of the vpc you will create if vpc_id not set
	CidrBlock                string `mapstructure:"cidr_block" required:"false"` // 10.0.0.0/16(default), 172.16.0.0/12, 192.168.0.0/16
	// Specify cider block of the subnet you will create if
    // subnet_id not set
	SubnectCidrBlock         string `mapstructure:"subnect_cidr_block" required:"false"`
	InternetChargeType       string `mapstructure:"internet_charge_type"`
	// Max bandwidth out your cvm will be launched by(in MB).
    // values can be set between 1 ~ 100.
	InternetMaxBandwidthOut  int64  `mapstructure:"internet_max_bandwidth_out" required:"false"`
	// Specify security group your cvm will be launched by.
	SecurityGroupId          string `mapstructure:"security_group_id" required:"false"`
	// Specify security name you will create if security_group_id not set.
	SecurityGroupName        string `mapstructure:"security_group_name" required:"false"`
	// userdata.
	UserData                 string `mapstructure:"user_data" required:"false"`
	// userdata file.
	UserDataFile             string `mapstructure:"user_data_file" required:"false"`
	// host name.
	HostName                 string `mapstructure:"host_name" required:"false"`

	// Communicator settings
	Comm         communicator.Config `mapstructure:",squash"`
	SSHPrivateIp bool                `mapstructure:"ssh_private_ip"`
}

var ValidCBSType = []string{
	"LOCAL_BASIC", "LOCAL_SSD", "CLOUD_BASIC", "CLOUD_SSD", "CLOUD_PREMIUM",
}

func (cf *TencentCloudRunConfig) Prepare(ctx *interpolate.Context) []error {
	if cf.Comm.SSHKeyPairName == "" && cf.Comm.SSHTemporaryKeyPairName == "" &&
		cf.Comm.SSHPrivateKeyFile == "" && cf.Comm.SSHPassword == "" && cf.Comm.WinRMPassword == "" {
		//tencentcloud support key pair name length max to 25
		cf.Comm.SSHTemporaryKeyPairName = fmt.Sprintf("packer_%s", uuid.TimeOrderedUUID()[:8])
	}

	errs := cf.Comm.Prepare(ctx)
	if cf.SourceImageId == "" {
		errs = append(errs, errors.New("source_image_id must be specified"))
	}

	if !CheckResourceIdFormat("img", cf.SourceImageId) {
		errs = append(errs, errors.New("source_image_id wrong format"))
	}

	if cf.InstanceType == "" {
		errs = append(errs, errors.New("instance_type must be specified"))
	}

	if cf.UserData != "" && cf.UserDataFile != "" {
		errs = append(errs, errors.New("only one of user_data or user_data_file can be specified"))
	} else if cf.UserDataFile != "" {
		if _, err := os.Stat(cf.UserDataFile); err != nil {
			errs = append(errs, errors.New("user_data_file not exist"))
		}
	}

	if (cf.VpcId != "" || cf.CidrBlock != "") && cf.SubnetId == "" && cf.SubnectCidrBlock == "" {
		errs = append(errs, errors.New("if vpc cidr_block is specified, then "+
			"subnet_cidr_block must also be specified."))
	}

	if cf.VpcId == "" {
		if cf.VpcName == "" {
			cf.VpcName = fmt.Sprintf("packer_%s", uuid.TimeOrderedUUID())
		}
		if cf.CidrBlock == "" {
			cf.CidrBlock = "10.0.0.0/16"
		}
		if cf.SubnetId != "" {
			errs = append(errs, errors.New("can't set subnet_id without set vpc_id"))
		}
	}
	if cf.SubnetId == "" {
		if cf.SubnetName == "" {
			cf.SubnetName = fmt.Sprintf("packer_%s", uuid.TimeOrderedUUID())
		}
		if cf.SubnectCidrBlock == "" {
			cf.SubnectCidrBlock = "10.0.8.0/24"
		}
	}

	if cf.SecurityGroupId == "" && cf.SecurityGroupName == "" {
		cf.SecurityGroupName = fmt.Sprintf("packer_%s", uuid.TimeOrderedUUID())
	}

	if cf.DiskType != "" && !checkDiskType(cf.DiskType) {
		errs = append(errs, errors.New(fmt.Sprintf("specified disk_type(%s) is invalid", cf.DiskType)))
	} else if cf.DiskType == "" {
		cf.DiskType = "CLOUD_BASIC"
	}

	if cf.DiskSize <= 0 {
		cf.DiskSize = 50
	}

	if cf.AssociatePublicIpAddress && cf.InternetMaxBandwidthOut <= 0 {
		cf.InternetMaxBandwidthOut = 1
	}

	if cf.InstanceName == "" {
		cf.InstanceName = fmt.Sprintf("packer-%s", uuid.TimeOrderedUUID())
	}

	if cf.HostName == "" {
		cf.HostName = cf.InstanceName[:15]
	}

	return errs
}

func checkDiskType(diskType string) bool {
	for _, valid := range ValidCBSType {
		if valid == diskType {
			return true
		}
	}
	return false
}

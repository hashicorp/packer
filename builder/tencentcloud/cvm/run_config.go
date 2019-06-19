package cvm

import (
	"fmt"
	"os"

	"github.com/hashicorp/packer/common/uuid"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/template/interpolate"
	"github.com/pkg/errors"
)

type tencentCloudDataDisk struct {
	DiskType   string `mapstructure:"disk_type"`
	DiskSize   int64  `mapstructure:"disk_size"`
	SnapshotId string `mapstructure:"disk_snapshot_id"`
}

type TencentCloudRunConfig struct {
	AssociatePublicIpAddress bool                   `mapstructure:"associate_public_ip_address"`
	SourceImageId            string                 `mapstructure:"source_image_id"`
	InstanceType             string                 `mapstructure:"instance_type"`
	InstanceName             string                 `mapstructure:"instance_name"`
	DiskType                 string                 `mapstructure:"disk_type"`
	DiskSize                 int64                  `mapstructure:"disk_size"`
	DataDisks                []tencentCloudDataDisk `mapstructure:"data_disks"`
	VpcId                    string                 `mapstructure:"vpc_id"`
	VpcName                  string                 `mapstructure:"vpc_name"`
	VpcIp                    string                 `mapstructure:"vpc_ip"`
	SubnetId                 string                 `mapstructure:"subnet_id"`
	SubnetName               string                 `mapstructure:"subnet_name"`
	CidrBlock                string                 `mapstructure:"cidr_block"` // 10.0.0.0/16(default), 172.16.0.0/12, 192.168.0.0/16
	SubnectCidrBlock         string                 `mapstructure:"subnect_cidr_block"`
	InternetChargeType       string                 `mapstructure:"internet_charge_type"`
	InternetMaxBandwidthOut  int64                  `mapstructure:"internet_max_bandwidth_out"`
	SecurityGroupId          string                 `mapstructure:"security_group_id"`
	SecurityGroupName        string                 `mapstructure:"security_group_name"`
	UserData                 string                 `mapstructure:"user_data"`
	UserDataFile             string                 `mapstructure:"user_data_file"`
	HostName                 string                 `mapstructure:"host_name"`
	RunTags                  map[string]string      `mapstructure:"run_tags"`

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

	if cf.RunTags == nil {
		cf.RunTags = make(map[string]string)
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

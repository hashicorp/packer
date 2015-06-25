package common

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/mitchellh/packer/template/interpolate"
)

// BlockDevice
type BlockDevice struct {
	DeleteOnTermination bool   `mapstructure:"delete_on_termination"`
	DeviceName          string `mapstructure:"device_name"`
	Encrypted           bool   `mapstructure:"encrypted"`
	IOPS                int64  `mapstructure:"iops"`
	NoDevice            bool   `mapstructure:"no_device"`
	SnapshotId          string `mapstructure:"snapshot_id"`
	VirtualName         string `mapstructure:"virtual_name"`
	VolumeType          string `mapstructure:"volume_type"`
	VolumeSize          int64  `mapstructure:"volume_size"`
}

type BlockDevices struct {
	AMIMappings    []BlockDevice `mapstructure:"ami_block_device_mappings"`
	LaunchMappings []BlockDevice `mapstructure:"launch_block_device_mappings"`
}

func buildBlockDevices(b []BlockDevice) []*ec2.BlockDeviceMapping {
	var blockDevices []*ec2.BlockDeviceMapping

	for _, blockDevice := range b {
		ebsBlockDevice := &ec2.EBSBlockDevice{
			VolumeType:          aws.String(blockDevice.VolumeType),
			VolumeSize:          aws.Long(blockDevice.VolumeSize),
			DeleteOnTermination: aws.Boolean(blockDevice.DeleteOnTermination),
		}

		// IOPS is only valid for SSD Volumes
		if blockDevice.VolumeType != "" && blockDevice.VolumeType != "standard" && blockDevice.VolumeType != "gp2" {
			ebsBlockDevice.IOPS = aws.Long(blockDevice.IOPS)
		}

		// You cannot specify Encrypted if you specify a Snapshot ID
		if blockDevice.SnapshotId != "" {
			ebsBlockDevice.SnapshotID = aws.String(blockDevice.SnapshotId)
		} else if blockDevice.Encrypted {
			ebsBlockDevice.Encrypted = aws.Boolean(blockDevice.Encrypted)
		}

		mapping := &ec2.BlockDeviceMapping{
			EBS:         ebsBlockDevice,
			DeviceName:  aws.String(blockDevice.DeviceName),
			VirtualName: aws.String(blockDevice.VirtualName),
		}

		if blockDevice.NoDevice {
			mapping.NoDevice = aws.String("")
		}

		blockDevices = append(blockDevices, mapping)
	}
	return blockDevices
}

func (b *BlockDevices) Prepare(ctx *interpolate.Context) []error {
	return nil
}

func (b *BlockDevices) BuildAMIDevices() []*ec2.BlockDeviceMapping {
	return buildBlockDevices(b.AMIMappings)
}

func (b *BlockDevices) BuildLaunchDevices() []*ec2.BlockDeviceMapping {
	return buildBlockDevices(b.LaunchMappings)
}

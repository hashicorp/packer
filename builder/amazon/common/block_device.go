package common

import (
	"fmt"

	"github.com/awslabs/aws-sdk-go/aws"
	"github.com/awslabs/aws-sdk-go/service/ec2"
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
			SnapshotID:          &blockDevice.SnapshotId,
			Encrypted:           &blockDevice.Encrypted,
			IOPS:                &blockDevice.IOPS,
			VolumeType:          &blockDevice.VolumeType,
			VolumeSize:          &blockDevice.VolumeSize,
			DeleteOnTermination: &blockDevice.DeleteOnTermination,
		}
		mapping := &ec2.BlockDeviceMapping{
			EBS:         ebsBlockDevice,
			DeviceName:  &blockDevice.DeviceName,
			VirtualName: &blockDevice.VirtualName,
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

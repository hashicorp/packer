package common

import (
	"strings"

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
		mapping := &ec2.BlockDeviceMapping{
			DeviceName: aws.String(blockDevice.DeviceName),
		}

		if blockDevice.NoDevice {
			mapping.NoDevice = aws.String("")
		} else if blockDevice.VirtualName != "" {
			if strings.HasPrefix(blockDevice.VirtualName, "ephemeral") {
				mapping.VirtualName = aws.String(blockDevice.VirtualName)
			}
		} else {
			ebsBlockDevice := &ec2.EbsBlockDevice{
				DeleteOnTermination: aws.Bool(blockDevice.DeleteOnTermination),
			}

			if blockDevice.VolumeType != "" {
				ebsBlockDevice.VolumeType = aws.String(blockDevice.VolumeType)
			}

			if blockDevice.VolumeSize > 0 {
				ebsBlockDevice.VolumeSize = aws.Int64(blockDevice.VolumeSize)
			}

			// IOPS is only valid for io1 type
			if blockDevice.VolumeType == "io1" {
				ebsBlockDevice.Iops = aws.Int64(blockDevice.IOPS)
			}

			// You cannot specify Encrypted if you specify a Snapshot ID
			if blockDevice.SnapshotId != "" {
				ebsBlockDevice.SnapshotId = aws.String(blockDevice.SnapshotId)
			} else if blockDevice.Encrypted {
				ebsBlockDevice.Encrypted = aws.Bool(blockDevice.Encrypted)
			}

			mapping.Ebs = ebsBlockDevice
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

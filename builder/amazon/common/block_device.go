package common

import (
	"github.com/mitchellh/goamz/ec2"
)

// BlockDevice
type BlockDevice struct {
	DeleteOnTermination bool   `mapstructure:"delete_on_termination"`
	DeviceName          string `mapstructure:"device_name"`
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

func buildBlockDevices(b []BlockDevice) []ec2.BlockDeviceMapping {
	var blockDevices []ec2.BlockDeviceMapping

	for _, blockDevice := range b {
		blockDevices = append(blockDevices, ec2.BlockDeviceMapping{
			DeviceName:          blockDevice.DeviceName,
			VirtualName:         blockDevice.VirtualName,
			SnapshotId:          blockDevice.SnapshotId,
			VolumeType:          blockDevice.VolumeType,
			VolumeSize:          blockDevice.VolumeSize,
			DeleteOnTermination: blockDevice.DeleteOnTermination,
			IOPS:                blockDevice.IOPS,
			NoDevice:            blockDevice.NoDevice,
		})
	}
	return blockDevices
}

func (b *BlockDevices) BuildAMIDevices() []ec2.BlockDeviceMapping {
	return buildBlockDevices(b.AMIMappings)
}

func (b *BlockDevices) BuildLaunchDevices() []ec2.BlockDeviceMapping {
	return buildBlockDevices(b.LaunchMappings)
}

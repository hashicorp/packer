package common

import (
	"fmt"

	"github.com/mitchellh/goamz/ec2"
	"github.com/mitchellh/packer/packer"
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
			Encrypted:           blockDevice.Encrypted,
		})
	}
	return blockDevices
}

func (b *BlockDevices) Prepare(t *packer.ConfigTemplate) []error {
	if t == nil {
		var err error
		t, err = packer.NewConfigTemplate()
		if err != nil {
			return []error{err}
		}
	}

	lists := map[string][]BlockDevice{
		"ami_block_device_mappings":    b.AMIMappings,
		"launch_block_device_mappings": b.LaunchMappings,
	}

	var errs []error
	for outer, bds := range lists {
		for i := 0; i < len(bds); i++ {
			templates := map[string]*string{
				"device_name":  &bds[i].DeviceName,
				"snapshot_id":  &bds[i].SnapshotId,
				"virtual_name": &bds[i].VirtualName,
				"volume_type":  &bds[i].VolumeType,
			}

			errs := make([]error, 0)
			for n, ptr := range templates {
				var err error
				*ptr, err = t.Process(*ptr, nil)
				if err != nil {
					errs = append(
						errs, fmt.Errorf(
							"Error processing %s[%d].%s: %s",
							outer, i, n, err))
				}
			}
		}
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

func (b *BlockDevices) BuildAMIDevices() []ec2.BlockDeviceMapping {
	return buildBlockDevices(b.AMIMappings)
}

func (b *BlockDevices) BuildLaunchDevices() []ec2.BlockDeviceMapping {
	return buildBlockDevices(b.LaunchMappings)
}

package ebsvolume

import (
	awscommon "github.com/mitchellh/packer/builder/amazon/common"
)

type BlockDevice struct {
	awscommon.BlockDevice `mapstructure:"-,squash"`
	Tags                  map[string]string `mapstructure:"tags"`
}

func commonBlockDevices(mappings []BlockDevice) awscommon.BlockDevices {
	result := make([]awscommon.BlockDevice, len(mappings))
	for i, mapping := range mappings {
		result[i] = mapping.BlockDevice
	}

	return awscommon.BlockDevices{
		LaunchBlockDevices: awscommon.LaunchBlockDevices{
			LaunchMappings: result,
		},
	}
}

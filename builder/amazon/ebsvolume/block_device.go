package ebsvolume

import (
	awscommon "github.com/mitchellh/packer/builder/amazon/common"
	"github.com/mitchellh/packer/template/interpolate"
)

type BlockDevice struct {
	awscommon.BlockDevice `mapstructure:"-,squash"`
	Tags                  map[string]string `mapstructure:"tags"`
}

func commonBlockDevices(mappings []BlockDevice, ctx *interpolate.Context) (awscommon.BlockDevices, error) {
	result := make([]awscommon.BlockDevice, len(mappings))

	for i, mapping := range mappings {
		interpolateBlockDev, err := interpolate.RenderInterface(&mapping.BlockDevice, ctx)
		if err != nil {
			return awscommon.BlockDevices{}, err
		}
		result[i] = *interpolateBlockDev.(*awscommon.BlockDevice)
	}

	return awscommon.BlockDevices{
		LaunchBlockDevices: awscommon.LaunchBlockDevices{
			LaunchMappings: result,
		},
	}, nil
}

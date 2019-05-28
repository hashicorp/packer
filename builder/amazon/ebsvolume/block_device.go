package ebsvolume

import (
	awscommon "github.com/hashicorp/packer/builder/amazon/common"
	"github.com/hashicorp/packer/template/interpolate"
)

type BlockDevice struct {
	awscommon.BlockDevice `mapstructure:"-,squash"`
	// Tags applied to the AMI. This is a
    // template engine, see Build template
    // data for more information.
	Tags                  awscommon.TagMap `mapstructure:"tags" required:"false"`
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

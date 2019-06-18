//go:generate struct-markdown

package ebsvolume

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	awscommon "github.com/hashicorp/packer/builder/amazon/common"
	"github.com/hashicorp/packer/template/interpolate"
)

type BlockDevice struct {
	awscommon.BlockDevice `mapstructure:",squash"`

	OmitFromArtifact bool `mapstructure:"omit_from_artifact"`
	// Tags applied to the AMI. This is a
	// template engine, see Build template
	// data for more information.
	Tags awscommon.TagMap `mapstructure:"tags" required:"false"`
}

type BlockDevices []BlockDevice

func (bds BlockDevices) BuildEC2BlockDeviceMappings() []*ec2.BlockDeviceMapping {
	var blockDevices []*ec2.BlockDeviceMapping

	for _, blockDevice := range bds {
		blockDevices = append(blockDevices, blockDevice.BuildEC2BlockDeviceMapping())
	}
	return blockDevices
}

func (bds BlockDevices) Prepare(ctx *interpolate.Context) (errs []error) {
	for _, block := range bds {
		if err := block.Prepare(ctx); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func (b BlockDevices) GetOmissions() map[string]bool {
	omitMap := make(map[string]bool)

	for _, blockDevice := range b {
		omitMap[blockDevice.DeviceName] = blockDevice.OmitFromArtifact
	}

	return omitMap
}

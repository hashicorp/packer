//go:generate struct-markdown

package ebsvolume

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	awscommon "github.com/hashicorp/packer/builder/amazon/common"
	"github.com/hashicorp/packer/template/interpolate"
)

type BlockDevice struct {
	awscommon.BlockDevice `mapstructure:",squash"`
	// Tags to apply to the volume. These are retained after the builder
	// completes. This is a [template engine](/docs/templates/engine.html), see
	// [Build template data](#build-template-data) for more information.
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

//go:generate struct-markdown

package ebsvolume

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	awscommon "github.com/hashicorp/packer/builder/amazon/common"
	"github.com/hashicorp/packer/hcl2template"
	"github.com/hashicorp/packer/template/interpolate"
)

type BlockDevice struct {
	awscommon.BlockDevice `mapstructure:",squash"`
	// Tags to apply to the volume. These are retained after the builder
	// completes. This is a [template engine](/docs/templates/engine.html), see
	// [Build template data](#build-template-data) for more information.
	Tags awscommon.TagMap `mapstructure:"tags" required:"false"`
	// Same as [`tags`](#tags) but defined as a singular repeatable block
	// containing a key and a value field. In HCL2 mode the
	// [`dynamic_block`](https://packer.io/docs/configuration/from-1.5/expressions.html#dynamic-blocks)
	// will allow you to create those programatically.
	Tag hcl2template.KeyValues `mapstructure:"tag" required:"false"`
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

		for _, s := range []struct {
			tagMap awscommon.TagMap
			kvs    hcl2template.KeyValues
		}{
			{block.Tags, block.Tag},
		} {
			errs = append(errs, s.kvs.CopyOn(s.tagMap)...)
		}

		if err := block.Prepare(ctx); err != nil {
			errs = append(errs, err)
		}

	}
	return errs
}

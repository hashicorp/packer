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
	// Key/value pair tags to apply to the volume. These are retained after the builder
	// completes. This is a [template engine](/docs/templates/engine), see
	// [Build template data](#build-template-data) for more information.
	Tags map[string]string `mapstructure:"tags" required:"false"`
	// Same as [`tags`](#tags) but defined as a singular repeatable block
	// containing a `key` and a `value` field. In HCL2 mode the
	// [`dynamic_block`](/docs/configuration/from-1.5/expressions#dynamic-blocks)
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

		errs = append(errs, block.Tag.CopyOn(&block.Tags)...)

		if err := block.Prepare(ctx); err != nil {
			errs = append(errs, err)
		}

	}
	return errs
}

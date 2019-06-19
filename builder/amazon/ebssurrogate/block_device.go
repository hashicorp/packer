//go:generate struct-markdown

package ebssurrogate

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	awscommon "github.com/hashicorp/packer/builder/amazon/common"
	"github.com/hashicorp/packer/template/interpolate"
)

type BlockDevice struct {
	awscommon.BlockDevice `mapstructure:",squash"`

	// If true, this block device will not be snapshotted and the created AMI
	// will not contain block device mapping information for this volume. If
	// false, the block device will be mapped into the final created AMI. Set
	// this option to true if you need a block device mounted in the surrogate
	// AMI but not in the final created AMI.
	OmitFromArtifact bool `mapstructure:"omit_from_artifact"`
}

type BlockDevices []BlockDevice

func (bds BlockDevices) Common() []awscommon.BlockDevice {
	res := []awscommon.BlockDevice{}
	for _, bd := range bds {
		res = append(res, bd.BlockDevice)
	}
	return res
}

func (bds BlockDevices) BuildEC2BlockDeviceMappings() []*ec2.BlockDeviceMapping {
	var blockDevices []*ec2.BlockDeviceMapping

	for _, blockDevice := range bds {
		blockDevices = append(blockDevices, blockDevice.BuildEC2BlockDeviceMapping())
	}
	return blockDevices
}

func (blockDevice BlockDevice) BuildEC2BlockDeviceMapping() *ec2.BlockDeviceMapping {

	mapping := &ec2.BlockDeviceMapping{
		DeviceName: aws.String(blockDevice.DeviceName),
	}

	if blockDevice.NoDevice {
		mapping.NoDevice = aws.String("")
		return mapping
	} else if blockDevice.VirtualName != "" {
		if strings.HasPrefix(blockDevice.VirtualName, "ephemeral") {
			mapping.VirtualName = aws.String(blockDevice.VirtualName)
		}
		return mapping
	}

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
	}
	ebsBlockDevice.Encrypted = blockDevice.Encrypted

	mapping.Ebs = ebsBlockDevice

	return mapping
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

//go:generate struct-markdown

package chroot

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	awscommon "github.com/hashicorp/packer/builder/amazon/common"
	"github.com/hashicorp/packer/template/interpolate"
)

type BlockDevice struct {
	awscommon.BlockDevice `mapstructure:",squash"`
	// ID, alias or ARN of the KMS key to use for boot volume encryption. This
	// only applies to the main region, other regions where the AMI will be
	// copied will be encrypted by the default EBS KMS key. For valid formats
	// see KmsKeyId in the [AWS API docs -
	// CopyImage](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_CopyImage.html)
	// This field is validated by Packer, when using an alias, you will have to
	// prefix kms_key_id with alias/.
	KmsKeyId string `mapstructure:"kms_key_id" required:"false"`
}

type BlockDevices []BlockDevice

func (bds BlockDevices) BuildEC2BlockDeviceMappings() []*ec2.BlockDeviceMapping {
	var blockDevices []*ec2.BlockDeviceMapping

	for _, blockDevice := range bds {
		blockDevices = append(blockDevices, blockDevice.BuildEC2BlockDeviceMapping())
	}
	return blockDevices
}

func (blockDevice BlockDevice) BuildEC2BlockDeviceMapping() *ec2.BlockDeviceMapping {
	mapping := blockDevice.BlockDevice.BuildEC2BlockDeviceMapping()

	if blockDevice.KmsKeyId != "" {
		mapping.Ebs.KmsKeyId = aws.String(blockDevice.KmsKeyId)
	}
	return mapping
}

func (b *BlockDevice) Prepare(ctx *interpolate.Context) error {
	if b.DeviceName == "" {
		return fmt.Errorf("The `device_name` must be specified " +
			"for every device in the block device mapping.")
	}
	// Warn that encrypted must be true when setting kms_key_id
	if b.KmsKeyId != "" && b.Encrypted != nil && *b.Encrypted == false {
		return fmt.Errorf("The device %v, must also have `encrypted: "+
			"true` when setting a kms_key_id.", b.DeviceName)
	}

	return nil
}

func (bds BlockDevices) Prepare(ctx *interpolate.Context) (errs []error) {
	for _, block := range bds {
		if err := block.Prepare(ctx); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

//go:generate struct-markdown
//go:generate mapstructure-to-hcl2 -type BlockDevice

package common

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/template/interpolate"
)

// These will be attached when booting a new instance from your AMI. Your
// options here may vary depending on the type of VM you use.
//
// Example use case:
//
// The following mapping will tell Packer to encrypt the root volume of the
// build instance at launch using a specific non-default kms key:
//
// ``` json
// "[{
//		"device_name": "/dev/sda1",
//		"encrypted": true,
//		"kms_key_id": "1a2b3c4d-5e6f-1a2b-3c4d-5e6f1a2b3c4d"
// }]
// ```
//
// Documentation for Block Devices Mappings can be found here:
// https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/block-device-mapping-concepts.html
//
type BlockDevice struct {
	// Indicates whether the EBS volume is deleted on instance termination.
	// Default false. NOTE: If this value is not explicitly set to true and
	// volumes are not cleaned up by an alternative method, additional volumes
	// will accumulate after every build.
	DeleteOnTermination bool `mapstructure:"delete_on_termination" required:"false"`
	// The device name exposed to the instance (for example, /dev/sdh or xvdh).
	// Required for every device in the block device mapping.
	DeviceName string `mapstructure:"device_name" required:"false"`
	// Indicates whether or not to encrypt the volume. By default, Packer will
	// keep the encryption setting to what it was in the source image. Setting
	// false will result in an unencrypted device, and true will result in an
	// encrypted one.
	Encrypted config.Trilean `mapstructure:"encrypted" required:"false"`
	// The number of I/O operations per second (IOPS) that the volume supports.
	// See the documentation on
	// [IOPs](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_EbsBlockDevice.html)
	// for more information
	IOPS int64 `mapstructure:"iops" required:"false"`
	// Suppresses the specified device included in the block device mapping of
	// the AMI.
	NoDevice bool `mapstructure:"no_device" required:"false"`
	// The ID of the snapshot.
	SnapshotId string `mapstructure:"snapshot_id" required:"false"`
	// The virtual device name. See the documentation on Block Device Mapping
	// for more information.
	VirtualName string `mapstructure:"virtual_name" required:"false"`
	// The volume type. gp2 for General Purpose (SSD) volumes, io1 for
	// Provisioned IOPS (SSD) volumes, st1 for Throughput Optimized HDD, sc1
	// for Cold HDD, and standard for Magnetic volumes.
	VolumeType string `mapstructure:"volume_type" required:"false"`
	// The size of the volume, in GiB. Required if not specifying a
	// snapshot_id.
	VolumeSize int64 `mapstructure:"volume_size" required:"false"`
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
	ebsBlockDevice.Encrypted = blockDevice.Encrypted.ToBoolPointer()

	if blockDevice.KmsKeyId != "" {
		ebsBlockDevice.KmsKeyId = aws.String(blockDevice.KmsKeyId)
	}

	mapping.Ebs = ebsBlockDevice

	return mapping
}

func (b *BlockDevice) Prepare(ctx *interpolate.Context) error {
	if b.DeviceName == "" {
		return fmt.Errorf("The `device_name` must be specified " +
			"for every device in the block device mapping.")
	}

	// Warn that encrypted must be true or nil when setting kms_key_id
	if b.KmsKeyId != "" && b.Encrypted.False() {
		return fmt.Errorf("The device %v, must also have `encrypted: "+
			"true` when setting a kms_key_id.", b.DeviceName)
	}

	_, err := interpolate.RenderInterface(&b, ctx)
	return err
}

func (bds BlockDevices) Prepare(ctx *interpolate.Context) (errs []error) {
	for _, block := range bds {
		if err := block.Prepare(ctx); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

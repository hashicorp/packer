//go:generate struct-markdown

package common

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/packer/template/interpolate"
)

// BlockDevice
type BlockDevice struct {
	// Indicates whether the EBS volume is
    // deleted on instance termination. Default false. NOTE: If this
    // value is not explicitly set to true and volumes are not cleaned up by
    // an alternative method, additional volumes will accumulate after every
    // build.
	DeleteOnTermination bool   `mapstructure:"delete_on_termination" required:"false"`
	// The device name exposed to the instance (for
    // example, /dev/sdh or xvdh). Required for every device in the block
    // device mapping.
	DeviceName          string `mapstructure:"device_name" required:"false"`
	// Indicates whether or not to encrypt the volume.
    // By default, Packer will keep the encryption setting to what it was in
    // the source image. Setting false will result in an unencrypted device,
    // and true will result in an encrypted one.
	Encrypted           *bool  `mapstructure:"encrypted" required:"false"`
	// The number of I/O operations per second (IOPS) that
    // the volume supports. See the documentation on
    // IOPs
    // for more information
	IOPS                int64  `mapstructure:"iops" required:"false"`
	// Suppresses the specified device included in the
    // block device mapping of the AMI.
	NoDevice            bool   `mapstructure:"no_device" required:"false"`
	// The ID of the snapshot.
	SnapshotId          string `mapstructure:"snapshot_id" required:"false"`
	// The virtual device name. See the
    // documentation on Block Device
    // Mapping
    // for more information.
	VirtualName         string `mapstructure:"virtual_name" required:"false"`
	// The volume type. gp2 for General Purpose
    // (SSD) volumes, io1 for Provisioned IOPS (SSD) volumes, st1 for
    // Throughput Optimized HDD, sc1 for Cold HDD, and standard for
    // Magnetic volumes.
	VolumeType          string `mapstructure:"volume_type" required:"false"`
	// The size of the volume, in GiB. Required if
    // not specifying a snapshot_id.
	VolumeSize          int64  `mapstructure:"volume_size" required:"false"`
	// ID, alias or ARN of the KMS key to use for boot
    // volume encryption. This only applies to the main region, other regions
    // where the AMI will be copied will be encrypted by the default EBS KMS key.
    // For valid formats see KmsKeyId in the AWS API docs -
    // CopyImage.
    // This field is validated by Packer, when using an alias, you will have to
    // prefix kms_key_id with alias/.
	KmsKeyId            string `mapstructure:"kms_key_id" required:"false"`
	// ebssurrogate only
	OmitFromArtifact bool `mapstructure:"omit_from_artifact"`
}

type BlockDevices struct {
	AMIBlockDevices    `mapstructure:",squash"`
	LaunchBlockDevices `mapstructure:",squash"`
}

type AMIBlockDevices struct {
	// Add one or
    // more block device
    // mappings
    // to the AMI. These will be attached when booting a new instance from your
    // AMI. If this field is populated, and you are building from an existing source image,
    // the block device mappings in the source image will be overwritten. This means you
    // must have a block device mapping entry for your root volume, root_volume_size,
    // and root_device_name. `Your options here may vary depending on the type of VM
    // you use. The block device mappings allow for the following configuration:
	AMIMappings []BlockDevice `mapstructure:"ami_block_device_mappings" required:"false"`
}

type LaunchBlockDevices struct {
	// Add one
    // or more block devices before the Packer build starts. If you add instance
    // store volumes or EBS volumes in addition to the root device volume, the
    // created AMI will contain block device mapping information for those
    // volumes. Amazon creates snapshots of the source instance's root volume and
    // any other EBS volumes described here. When you launch an instance from this
    // new AMI, the instance automatically launches with these additional volumes,
    // and will restore them from snapshots taken from the source instance.
	LaunchMappings []BlockDevice `mapstructure:"launch_block_device_mappings" required:"false"`
}

func buildBlockDevices(b []BlockDevice) []*ec2.BlockDeviceMapping {
	var blockDevices []*ec2.BlockDeviceMapping

	for _, blockDevice := range b {
		mapping := &ec2.BlockDeviceMapping{
			DeviceName: aws.String(blockDevice.DeviceName),
		}

		if blockDevice.NoDevice {
			mapping.NoDevice = aws.String("")
		} else if blockDevice.VirtualName != "" {
			if strings.HasPrefix(blockDevice.VirtualName, "ephemeral") {
				mapping.VirtualName = aws.String(blockDevice.VirtualName)
			}
		} else {
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

			if blockDevice.KmsKeyId != "" {
				ebsBlockDevice.KmsKeyId = aws.String(blockDevice.KmsKeyId)
			}

			mapping.Ebs = ebsBlockDevice
		}

		blockDevices = append(blockDevices, mapping)
	}
	return blockDevices
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

func (b *BlockDevices) Prepare(ctx *interpolate.Context) (errs []error) {
	for _, d := range b.AMIMappings {
		if err := d.Prepare(ctx); err != nil {
			errs = append(errs, fmt.Errorf("AMIMapping: %s", err.Error()))
		}
	}
	for _, d := range b.LaunchMappings {
		if err := d.Prepare(ctx); err != nil {
			errs = append(errs, fmt.Errorf("LaunchMapping: %s", err.Error()))
		}
	}
	return errs
}

func (b *AMIBlockDevices) BuildAMIDevices() []*ec2.BlockDeviceMapping {
	return buildBlockDevices(b.AMIMappings)
}

func (b *LaunchBlockDevices) BuildLaunchDevices() []*ec2.BlockDeviceMapping {
	return buildBlockDevices(b.LaunchMappings)
}

func (b *LaunchBlockDevices) GetOmissions() map[string]bool {
	omitMap := make(map[string]bool)

	for _, blockDevice := range b.LaunchMappings {
		omitMap[blockDevice.DeviceName] = blockDevice.OmitFromArtifact
	}

	return omitMap
}

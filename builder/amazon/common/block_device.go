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
	DeleteOnTermination bool   `mapstructure:"delete_on_termination"`
	DeviceName          string `mapstructure:"device_name"`
	Encrypted           bool   `mapstructure:"encrypted"`
	IOPS                int64  `mapstructure:"iops"`
	NoDevice            bool   `mapstructure:"no_device"`
	SnapshotId          string `mapstructure:"snapshot_id"`
	VirtualName         string `mapstructure:"virtual_name"`
	VolumeType          string `mapstructure:"volume_type"`
	VolumeSize          int64  `mapstructure:"volume_size"`
	KmsKeyId            string `mapstructure:"kms_key_id"`
}

type BlockDevices struct {
	AMIBlockDevices    `mapstructure:",squash"`
	LaunchBlockDevices `mapstructure:",squash"`
}

type AMIBlockDevices struct {
	AMIMappings []BlockDevice `mapstructure:"ami_block_device_mappings"`
}

type LaunchBlockDevices struct {
	LaunchMappings []BlockDevice `mapstructure:"launch_block_device_mappings"`
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
			} else if blockDevice.Encrypted {
				ebsBlockDevice.Encrypted = aws.Bool(blockDevice.Encrypted)
			}

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
	// Warn that encrypted must be true when setting kms_key_id
	if b.KmsKeyId != "" && b.Encrypted == false {
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

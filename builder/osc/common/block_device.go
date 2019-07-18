package common

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/packer/template/interpolate"
	"github.com/outscale/osc-go/oapi"
)

// BlockDevice
type BlockDevice struct {
	DeleteOnVmDeletion bool   `mapstructure:"delete_on_vm_deletion"`
	DeviceName         string `mapstructure:"device_name"`
	IOPS               int64  `mapstructure:"iops"`
	NoDevice           bool   `mapstructure:"no_device"`
	SnapshotId         string `mapstructure:"snapshot_id"`
	VirtualName        string `mapstructure:"virtual_name"`
	VolumeType         string `mapstructure:"volume_type"`
	VolumeSize         int64  `mapstructure:"volume_size"`
}

type BlockDevices struct {
	OMIBlockDevices    `mapstructure:",squash"`
	LaunchBlockDevices `mapstructure:",squash"`
}

type OMIBlockDevices struct {
	OMIMappings []BlockDevice `mapstructure:"omi_block_device_mappings"`
}

type LaunchBlockDevices struct {
	LaunchMappings []BlockDevice `mapstructure:"launch_block_device_mappings"`
}

func buildBlockDevicesImage(b []BlockDevice) []oapi.BlockDeviceMappingImage {
	var blockDevices []oapi.BlockDeviceMappingImage

	for _, blockDevice := range b {
		mapping := oapi.BlockDeviceMappingImage{
			DeviceName: blockDevice.DeviceName,
		}

		if blockDevice.VirtualName != "" {
			if strings.HasPrefix(blockDevice.VirtualName, "ephemeral") {
				mapping.VirtualDeviceName = blockDevice.VirtualName
			}
		} else {
			bsu := oapi.BsuToCreate{
				DeleteOnVmDeletion: aws.Bool(blockDevice.DeleteOnVmDeletion),
			}

			if blockDevice.VolumeType != "" {
				bsu.VolumeType = blockDevice.VolumeType
			}

			if blockDevice.VolumeSize > 0 {
				bsu.VolumeSize = blockDevice.VolumeSize
			}

			// IOPS is only valid for io1 type
			if blockDevice.VolumeType == "io1" {
				bsu.Iops = blockDevice.IOPS
			}

			if blockDevice.SnapshotId != "" {
				bsu.SnapshotId = blockDevice.SnapshotId
			}

			mapping.Bsu = bsu
		}

		blockDevices = append(blockDevices, mapping)
	}
	return blockDevices
}

func buildBlockDevicesVmCreation(b []BlockDevice) []oapi.BlockDeviceMappingVmCreation {
	log.Printf("[DEBUG] Launch Block Device %#v", b)

	var blockDevices []oapi.BlockDeviceMappingVmCreation

	for _, blockDevice := range b {
		mapping := oapi.BlockDeviceMappingVmCreation{
			DeviceName: blockDevice.DeviceName,
		}

		if blockDevice.NoDevice {
			mapping.NoDevice = ""
		} else if blockDevice.VirtualName != "" {
			if strings.HasPrefix(blockDevice.VirtualName, "ephemeral") {
				mapping.VirtualDeviceName = blockDevice.VirtualName
			}
		} else {
			bsu := oapi.BsuToCreate{
				DeleteOnVmDeletion: aws.Bool(blockDevice.DeleteOnVmDeletion),
			}

			if blockDevice.VolumeType != "" {
				bsu.VolumeType = blockDevice.VolumeType
			}

			if blockDevice.VolumeSize > 0 {
				bsu.VolumeSize = blockDevice.VolumeSize
			}

			// IOPS is only valid for io1 type
			if blockDevice.VolumeType == "io1" {
				bsu.Iops = blockDevice.IOPS
			}

			if blockDevice.SnapshotId != "" {
				bsu.SnapshotId = blockDevice.SnapshotId
			}

			mapping.Bsu = bsu
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
	return nil
}

func (b *BlockDevices) Prepare(ctx *interpolate.Context) (errs []error) {
	for _, d := range b.OMIMappings {
		if err := d.Prepare(ctx); err != nil {
			errs = append(errs, fmt.Errorf("OMIMapping: %s", err.Error()))
		}
	}
	for _, d := range b.LaunchMappings {
		if err := d.Prepare(ctx); err != nil {
			errs = append(errs, fmt.Errorf("LaunchMapping: %s", err.Error()))
		}
	}
	return errs
}

func (b *OMIBlockDevices) BuildOMIDevices() []oapi.BlockDeviceMappingImage {
	return buildBlockDevicesImage(b.OMIMappings)
}

func (b *LaunchBlockDevices) BuildLaunchDevices() []oapi.BlockDeviceMappingVmCreation {
	return buildBlockDevicesVmCreation(b.LaunchMappings)
}

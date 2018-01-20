package ecs

import (
	"fmt"

	"github.com/denverdino/aliyungo/ecs"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepMountAlicloudDisk struct {
}

func (s *stepMountAlicloudDisk) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*ecs.Client)
	config := state.Get("config").(Config)
	ui := state.Get("ui").(packer.Ui)
	instance := state.Get("instance").(*ecs.InstanceAttributesType)
	alicloudDiskDevices := config.ECSImagesDiskMappings
	if len(config.ECSImagesDiskMappings) == 0 {
		return multistep.ActionContinue
	}
	ui.Say("Mounting disks.")
	disks, _, err := client.DescribeDisks(&ecs.DescribeDisksArgs{InstanceId: instance.InstanceId,
		RegionId: instance.RegionId})
	if err != nil {
		err := fmt.Errorf("Error querying disks: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	for _, disk := range disks {
		if disk.Status == ecs.DiskStatusAvailable {
			if err := client.AttachDisk(&ecs.AttachDiskArgs{DiskId: disk.DiskId,
				InstanceId: instance.InstanceId,
				Device:     getDevice(&disk, alicloudDiskDevices),
			}); err != nil {
				err := fmt.Errorf("Error mounting disks: %s", err)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}
		}
	}
	for _, disk := range disks {
		if err := client.WaitForDisk(instance.RegionId, disk.DiskId, ecs.DiskStatusInUse, ALICLOUD_DEFAULT_SHORT_TIMEOUT); err != nil {
			err := fmt.Errorf("Timeout waiting for mount: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}
	ui.Say("Finished mounting disks.")
	return multistep.ActionContinue
}

func (s *stepMountAlicloudDisk) Cleanup(state multistep.StateBag) {

}

func getDevice(disk *ecs.DiskItemType, diskDevices []AlicloudDiskDevice) string {
	if disk.Device != "" {
		return disk.Device
	}
	for _, alicloudDiskDevice := range diskDevices {
		if alicloudDiskDevice.DiskName == disk.DiskName || alicloudDiskDevice.SnapshotId == disk.SourceSnapshotId {
			return alicloudDiskDevice.Device
		}
	}
	return ""
}

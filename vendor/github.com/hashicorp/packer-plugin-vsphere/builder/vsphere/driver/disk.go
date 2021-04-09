package driver

import (
	"errors"

	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/types"
)

type Disk struct {
	DiskSize            int64
	DiskEagerlyScrub    bool
	DiskThinProvisioned bool
	ControllerIndex     int
}

type StorageConfig struct {
	DiskControllerType []string // example: "scsi", "pvscsi", "nvme", "lsilogic"
	Storage            []Disk
}

func (c *StorageConfig) AddStorageDevices(existingDevices object.VirtualDeviceList) ([]types.BaseVirtualDeviceConfigSpec, error) {
	newDevices := object.VirtualDeviceList{}

	// Create new controller based on existing devices list and add it to the new devices list
	// to confirm creation
	var controllers []types.BaseVirtualController
	for _, controllerType := range c.DiskControllerType {
		var device types.BaseVirtualDevice
		var err error
		if controllerType == "nvme" {
			device, err = existingDevices.CreateNVMEController()
		} else {
			device, err = existingDevices.CreateSCSIController(controllerType)
		}
		if err != nil {
			return nil, err
		}
		existingDevices = append(existingDevices, device)
		newDevices = append(newDevices, device)
		controller, err := existingDevices.FindDiskController(existingDevices.Name(device))
		if err != nil {
			return nil, err
		}
		controllers = append(controllers, controller)
	}

	for _, dc := range c.Storage {
		disk := &types.VirtualDisk{
			VirtualDevice: types.VirtualDevice{
				Key: existingDevices.NewKey(),
				Backing: &types.VirtualDiskFlatVer2BackingInfo{
					DiskMode:        string(types.VirtualDiskModePersistent),
					ThinProvisioned: types.NewBool(dc.DiskThinProvisioned),
					EagerlyScrub:    types.NewBool(dc.DiskEagerlyScrub),
				},
			},
			CapacityInKB: dc.DiskSize * 1024,
		}

		existingDevices.AssignController(disk, controllers[dc.ControllerIndex])
		existingDevices = append(existingDevices, disk)
		newDevices = append(newDevices, disk)
	}

	return newDevices.ConfigSpec(types.VirtualDeviceConfigSpecOperationAdd)
}

func findDisk(devices object.VirtualDeviceList) (*types.VirtualDisk, error) {
	var disks []*types.VirtualDisk
	for _, device := range devices {
		switch d := device.(type) {
		case *types.VirtualDisk:
			disks = append(disks, d)
		}
	}

	switch len(disks) {
	case 0:
		return nil, errors.New("VM has no disks")
	case 1:
		return disks[0], nil
	}
	return nil, errors.New("VM has multiple disks")
}

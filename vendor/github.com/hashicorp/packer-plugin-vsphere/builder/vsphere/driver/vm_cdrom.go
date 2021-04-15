package driver

import (
	"errors"

	"github.com/vmware/govmomi/vim25/types"
)

var (
	ErrNoSataController = errors.New("no available SATA controller")
)

func (vm *VirtualMachineDriver) AddSATAController() error {
	sata := &types.VirtualAHCIController{}
	return vm.addDevice(sata)
}

func (vm *VirtualMachineDriver) FindSATAController() (*types.VirtualAHCIController, error) {
	l, err := vm.Devices()
	if err != nil {
		return nil, err
	}

	c := l.PickController((*types.VirtualAHCIController)(nil))
	if c == nil {
		return nil, ErrNoSataController
	}

	return c.(*types.VirtualAHCIController), nil
}

func (vm *VirtualMachineDriver) CreateCdrom(c *types.VirtualController) (*types.VirtualCdrom, error) {
	l, err := vm.Devices()
	if err != nil {
		return nil, err
	}

	device := &types.VirtualCdrom{}

	l.AssignController(device, c)

	device.Backing = &types.VirtualCdromAtapiBackingInfo{
		VirtualDeviceDeviceBackingInfo: types.VirtualDeviceDeviceBackingInfo{},
	}

	device.Connectable = &types.VirtualDeviceConnectInfo{
		AllowGuestControl: true,
		Connected:         true,
		StartConnected:    true,
	}

	return device, nil
}

func (vm *VirtualMachineDriver) RemoveCdroms() error {
	devices, err := vm.Devices()
	if err != nil {
		return err
	}
	cdroms := devices.SelectByType((*types.VirtualCdrom)(nil))
	if err = vm.RemoveDevice(true, cdroms...); err != nil {
		return err
	}

	sata := devices.SelectByType((*types.VirtualAHCIController)(nil))
	if err = vm.RemoveDevice(true, sata...); err != nil {
		return err
	}
	return nil
}

func (vm *VirtualMachineDriver) EjectCdroms() error {
	devices, err := vm.Devices()
	if err != nil {
		return err
	}
	cdroms := devices.SelectByType((*types.VirtualCdrom)(nil))
	for _, cd := range cdroms {
		c := cd.(*types.VirtualCdrom)
		c.Backing = &types.VirtualCdromRemotePassthroughBackingInfo{}
		c.Connectable = &types.VirtualDeviceConnectInfo{}
		err := vm.vm.EditDevice(vm.driver.ctx, c)
		if err != nil {
			return err
		}
	}

	return nil
}

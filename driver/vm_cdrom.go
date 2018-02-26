package driver

import (
	"github.com/vmware/govmomi/vim25/types"
	"errors"
)

func (vm *VirtualMachine) AddSATAController() error {
	sata := &types.VirtualAHCIController{}
	return vm.addDevice(sata)
}

func (vm *VirtualMachine) FindSATAController() (*types.VirtualAHCIController, error) {
	l, err := vm.Devices()
	if err != nil {
		return nil, err
	}

	c := l.PickController((*types.VirtualAHCIController)(nil))
	if c == nil {
		return nil, errors.New("no available SATA controller")
	}

	return c.(*types.VirtualAHCIController), nil
}

func (vm *VirtualMachine) CreateCdrom(c *types.VirtualAHCIController) (*types.VirtualCdrom, error) {
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

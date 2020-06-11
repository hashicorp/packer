package driver

import (
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/types"
	"golang.org/x/mobile/event/key"
)

type KeyInput struct {
	Message  string
	Scancode key.Code
	Alt      bool
	Ctrl     bool
	Shift    bool
}

func (vm *VirtualMachine) TypeOnKeyboard(spec types.UsbScanCodeSpec) (int32, error) {
	req := &types.PutUsbScanCodes{
		This: vm.vm.Reference(),
		Spec: spec,
	}

	resp, err := methods.PutUsbScanCodes(vm.driver.ctx, vm.driver.client.RoundTripper, req)
	if err != nil {
		return 0, err
	}

	return resp.Returnval, nil
}

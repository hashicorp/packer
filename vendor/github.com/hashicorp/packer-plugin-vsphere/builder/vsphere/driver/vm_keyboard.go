package driver

import (
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/types"
	"golang.org/x/mobile/event/key"
)

type KeyInput struct {
	Scancode key.Code
	Alt      bool
	Ctrl     bool
	Shift    bool
}

func (vm *VirtualMachineDriver) TypeOnKeyboard(input KeyInput) (int32, error) {
	var spec types.UsbScanCodeSpec

	spec.KeyEvents = append(spec.KeyEvents, types.UsbScanCodeSpecKeyEvent{
		UsbHidCode: int32(input.Scancode)<<16 | 7,
		Modifiers: &types.UsbScanCodeSpecModifierType{
			LeftControl: &input.Ctrl,
			LeftAlt:     &input.Alt,
			LeftShift:   &input.Shift,
		},
	})

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

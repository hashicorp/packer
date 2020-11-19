package common

import (
	"github.com/hashicorp/packer/builder/vsphere/driver"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

func CleanupVM(state multistep.StateBag) {
	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)
	_, destroy := state.GetOk("destroy_vm")
	if !cancelled && !halted && !destroy {
		return
	}

	ui := state.Get("ui").(packersdk.Ui)

	st := state.Get("vm")
	if st == nil {
		return
	}
	vm := st.(driver.VirtualMachine)

	ui.Say("Destroying VM...")
	err := vm.Destroy()
	if err != nil {
		ui.Error(err.Error())
	}
}

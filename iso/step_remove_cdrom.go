package iso

import (
	"fmt"

	"github.com/hashicorp/packer/packer"
	"github.com/jetbrains-infra/packer-builder-vsphere/driver"
	"github.com/mitchellh/multistep"
	"github.com/vmware/govmomi/vim25/types"
)

type StepRemoveCDRom struct{}

func (s *StepRemoveCDRom) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	vm := state.Get("vm").(*driver.VirtualMachine)

	devices, err := vm.Devices()
	if err != nil {
		ui.Error(fmt.Sprintf("error removing cdroms: %v", err))
		return multistep.ActionHalt
	}
	cdroms := devices.SelectByType((*types.VirtualCdrom)(nil))
	if err = vm.RemoveDevice(false, cdroms...); err != nil {
		ui.Error(fmt.Sprintf("error removing cdroms: %v", err))
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepRemoveCDRom) Cleanup(state multistep.StateBag) {}

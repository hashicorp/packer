package iso

import (
	"github.com/hashicorp/packer/packer"
	"github.com/jetbrains-infra/packer-builder-vsphere/driver"
	"github.com/mitchellh/multistep"
	"fmt"
	"github.com/vmware/govmomi/vim25/types"
)

type CDRomConfig struct {
	ISOPaths []string `mapstructure:"iso_paths"`
}

func (c *CDRomConfig) Prepare() []error {
	return nil
}

type StepAddCDRom struct {
	Config *CDRomConfig
}

func (s *StepAddCDRom) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	vm := state.Get("vm").(*driver.VirtualMachine)

	ui.Say("Adding CDRoms...")
	if err := vm.AddSATAController(); err != nil {
		state.Put("error", fmt.Errorf("error adding SATA controller: %v", err))
		return multistep.ActionHalt
	}

	for _, path := range s.Config.ISOPaths {
		if err := vm.AddCdrom(path); err != nil {
			state.Put("error", fmt.Errorf("error adding a cdrom: %v", err))
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (s *StepAddCDRom) Cleanup(state multistep.StateBag) {
	ui := state.Get("ui").(packer.Ui)
	vm := state.Get("vm").(*driver.VirtualMachine)

	devices, err := vm.Devices()
	if err != nil {
		ui.Error(fmt.Sprintf("error removing cdroms: %v", err))
	}
	cdroms := devices.SelectByType((*types.VirtualCdrom)(nil))
	if err = vm.RemoveDevice(false, cdroms...); err != nil {
		ui.Error(fmt.Sprintf("error removing cdroms: %v", err))
	}
}

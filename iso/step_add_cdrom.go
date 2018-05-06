package iso

import (
	"github.com/hashicorp/packer/packer"
	"github.com/jetbrains-infra/packer-builder-vsphere/driver"
	"github.com/hashicorp/packer/helper/multistep"
	"fmt"
	"context"
)

type CDRomConfig struct {
	ISOPaths []string `mapstructure:"iso_paths"`
}

type StepAddCDRom struct {
	Config *CDRomConfig
}

func (s *StepAddCDRom) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	vm := state.Get("vm").(*driver.VirtualMachine)

	ui.Say("Adding CD-ROM drives...")
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

func (s *StepAddCDRom) Cleanup(state multistep.StateBag) {}

package iso

import (
	"github.com/hashicorp/packer/packer"
	"github.com/jetbrains-infra/packer-builder-vsphere/driver"
	"github.com/mitchellh/multistep"
	"fmt"
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

	ui.Say("Adding CDRoms...")

	vm := state.Get("vm").(*driver.VirtualMachine)
	for _, path := range s.Config.ISOPaths {
		if err := vm.AddCdrom(path); err != nil {
			state.Put("error", fmt.Errorf("error adding a cdrom: %v", err))
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (s *StepAddCDRom) Cleanup(state multistep.StateBag) {
	// nothing
}

package iso

import (
	"github.com/hashicorp/packer/packer"
	"github.com/jetbrains-infra/packer-builder-vsphere/driver"
	"github.com/mitchellh/multistep"
)

type CDRomConfig struct {
	ISOPath string `mapstructure:"iso_path"`
}

func (c *CDRomConfig) Prepare() []error {
	var errs []error

	return errs
}

type StepAddCDRom struct {
	config *CDRomConfig
}

func (s *StepAddCDRom) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Adding CDRom ...")

	vm := state.Get("vm").(*driver.VirtualMachine)
	err := vm.AddCdrom(s.config.ISOPath)
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepAddCDRom) Cleanup(state multistep.StateBag) {
	// nothing
}

package main

import (
	"github.com/mitchellh/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/vmware/govmomi/object"
	"github.com/jetbrains-infra/packer-builder-vsphere/driver"
)

type StepConvertToTemplate struct{
	ConvertToTemplate bool
}

func (s *StepConvertToTemplate) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	d := state.Get("driver").(*driver.Driver)
	vm := state.Get("vm").(*object.VirtualMachine)

	if s.ConvertToTemplate {
		ui.Say("Convert VM into template...")
		err := d.ConvertToTemplate(vm)
		if err != nil {
			state.Put("error", err)
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (s *StepConvertToTemplate) Cleanup(state multistep.StateBag) {}

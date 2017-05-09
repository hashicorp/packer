package main

import (
	"github.com/mitchellh/multistep"
	"github.com/hashicorp/packer/packer"
	"strconv"
	"github.com/vmware/govmomi/vim25/types"
)

type StepConfigureHW struct{
	config *Config
}

func (s *StepConfigureHW) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	ui.Say("configuring virtual hardware...")

	var confSpec types.VirtualMachineConfigSpec
	// configure HW
	if s.config.Cpus != "" {
		cpus, err := strconv.Atoi(s.config.Cpus)
		if err != nil {
			state.Put("error", err)
			return multistep.ActionHalt
		}

		confSpec.NumCPUs = int32(cpus)
	}
	if s.config.Ram != "" {
		ram, err := strconv.Atoi(s.config.Ram)
		if err != nil {
			state.Put("error", err)
			return multistep.ActionHalt
		}

		confSpec.MemoryMB = int64(ram)
	}

	state.Put("confSpec", confSpec)

	return multistep.ActionContinue
}

func (s *StepConfigureHW) Cleanup(multistep.StateBag) {}

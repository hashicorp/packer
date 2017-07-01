package main

import (
	"github.com/mitchellh/multistep"
	"github.com/hashicorp/packer/packer"
	"strconv"
	"github.com/vmware/govmomi/vim25/types"
	"context"
	"github.com/vmware/govmomi/object"
)

type StepConfigureHW struct{
	config *Config
}

type ConfigParametersFlag struct {
	NumCPUsPtr  *int32
	MemoryMBPtr *int64
}

func (s *StepConfigureHW) Run(state multistep.StateBag) multistep.StepAction {
	vm := state.Get("vm").(*object.VirtualMachine)
	ctx := state.Get("ctx").(context.Context)

	var confSpec types.VirtualMachineConfigSpec
	parametersFlag := ConfigParametersFlag{}
	// configure HW
	if s.config.CPUs != "" {
		CPUs, err := strconv.Atoi(s.config.CPUs)
		if err != nil {
			state.Put("error", err)
			return multistep.ActionHalt
		}

		confSpec.NumCPUs = int32(CPUs)
		parametersFlag.NumCPUsPtr = &(confSpec.NumCPUs)
	}
	if s.config.RAM != "" {
		ram, err := strconv.Atoi(s.config.RAM)
		if err != nil {
			state.Put("error", err)
			return multistep.ActionHalt
		}

		confSpec.MemoryMB = int64(ram)
		parametersFlag.MemoryMBPtr = &(confSpec.MemoryMB)
	}

	ui := state.Get("ui").(packer.Ui)
	if parametersFlag != (ConfigParametersFlag{}) {
		ui.Say("configuring virtual hardware...")
		// Reconfigure hardware
		task, err := vm.Reconfigure(ctx, confSpec)
		if err != nil {
			state.Put("error", err)
			return multistep.ActionHalt
		}
		_, err = task.WaitForResult(ctx, nil)
		if err != nil {
			state.Put("error", err)
			return multistep.ActionHalt
		}
	} else {
		ui.Say("skipping the virtual hardware configration...")
	}

	return multistep.ActionContinue
}

func (s *StepConfigureHW) Cleanup(multistep.StateBag) {}

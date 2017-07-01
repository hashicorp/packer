package main

import (
	"github.com/mitchellh/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/vmware/govmomi/vim25/types"
	"context"
	"github.com/vmware/govmomi/object"
)

type HardwareConfig struct {
	CPUs           int32 `mapstructure:"CPUs"`
	CPUReservation int64 `mapstructure:"CPU_reservation"`
	CPULimit       int64 `mapstructure:"CPU_limit"`
	RAM            int64 `mapstructure:"RAM"`
	RAMReservation int64 `mapstructure:"RAM_reservation"`
}

type StepConfigureHardware struct {
	config *HardwareConfig
}

func (s *StepConfigureHardware) Run(state multistep.StateBag) multistep.StepAction {
	vm := state.Get("vm").(*object.VirtualMachine)
	ctx := state.Get("ctx").(context.Context)
	ui := state.Get("ui").(packer.Ui)

	if *s.config != (HardwareConfig{}) {
		ui.Say("Customizing hardware parameters...")

		var confSpec types.VirtualMachineConfigSpec
		confSpec.NumCPUs = s.config.CPUs
		confSpec.MemoryMB = s.config.RAM

		var cpuSpec types.ResourceAllocationInfo
		cpuSpec.Reservation = s.config.CPUReservation
		cpuSpec.Limit = s.config.CPULimit
		confSpec.CpuAllocation = &cpuSpec

		var ramSpec types.ResourceAllocationInfo
		ramSpec.Reservation = s.config.RAMReservation
		confSpec.MemoryAllocation = &ramSpec

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
	}

	return multistep.ActionContinue
}

func (s *StepConfigureHardware) Cleanup(multistep.StateBag) {}

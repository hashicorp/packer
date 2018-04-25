package clone

import (
	"github.com/hashicorp/packer/packer"
	"github.com/jetbrains-infra/packer-builder-vsphere/common"
	"github.com/jetbrains-infra/packer-builder-vsphere/driver"
	"github.com/hashicorp/packer/helper/multistep"
	"context"
)

type StepConfigureHardware struct {
	config *common.HardwareConfig
}

func (s *StepConfigureHardware) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	vm := state.Get("vm").(*driver.VirtualMachine)

	if *s.config != (common.HardwareConfig{}) {
		ui.Say("Customizing hardware parameters...")

		err := vm.Configure(&driver.HardwareConfig{
			CPUs:                s.config.CPUs,
			CPUReservation:      s.config.CPUReservation,
			CPULimit:            s.config.CPULimit,
			RAM:                 s.config.RAM,
			RAMReservation:      s.config.RAMReservation,
			RAMReserveAll:       s.config.RAMReserveAll,
			DiskSize:            s.config.DiskSize,
			NestedHV:            s.config.NestedHV,
			CpuHotAddEnabled:    s.config.CpuHotAddEnabled,
			MemoryHotAddEnabled: s.config.MemoryHotAddEnabled,
		})
		if err != nil {
			state.Put("error", err)
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (s *StepConfigureHardware) Cleanup(multistep.StateBag) {}

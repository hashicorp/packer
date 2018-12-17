package common

import (
	"context"
	"fmt"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/jetbrains-infra/packer-builder-vsphere/driver"
)

type HardwareConfig struct {
	CPUs             int32 `mapstructure:"CPUs"`
	CPUReservation   int64 `mapstructure:"CPU_reservation"`
	CPULimit         int64 `mapstructure:"CPU_limit"`
	CpuHotAddEnabled bool  `mapstructure:"CPU_hot_plug"`

	RAM                 int64 `mapstructure:"RAM"`
	RAMReservation      int64 `mapstructure:"RAM_reservation"`
	RAMReserveAll       bool  `mapstructure:"RAM_reserve_all"`
	MemoryHotAddEnabled bool  `mapstructure:"RAM_hot_plug"`

	VideoRAM int64 `mapstructure:"video_ram"`
	NestedHV bool  `mapstructure:"NestedHV"`
}

func (c *HardwareConfig) Prepare() []error {
	var errs []error

	if c.RAMReservation > 0 && c.RAMReserveAll != false {
		errs = append(errs, fmt.Errorf("'RAM_reservation' and 'RAM_reserve_all' cannot be used together"))
	}

	return errs
}

type StepConfigureHardware struct {
	Config *HardwareConfig
}

func (s *StepConfigureHardware) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	vm := state.Get("vm").(*driver.VirtualMachine)

	if *s.Config != (HardwareConfig{}) {
		ui.Say("Customizing hardware...")

		err := vm.Configure(&driver.HardwareConfig{
			CPUs:                s.Config.CPUs,
			CPUReservation:      s.Config.CPUReservation,
			CPULimit:            s.Config.CPULimit,
			RAM:                 s.Config.RAM,
			RAMReservation:      s.Config.RAMReservation,
			RAMReserveAll:       s.Config.RAMReserveAll,
			NestedHV:            s.Config.NestedHV,
			CpuHotAddEnabled:    s.Config.CpuHotAddEnabled,
			MemoryHotAddEnabled: s.Config.MemoryHotAddEnabled,
			VideoRAM:            s.Config.VideoRAM,
		})
		if err != nil {
			state.Put("error", err)
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (s *StepConfigureHardware) Cleanup(multistep.StateBag) {}

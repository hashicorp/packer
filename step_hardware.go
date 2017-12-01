package main

import (
	"github.com/mitchellh/multistep"
	"github.com/hashicorp/packer/packer"
	"fmt"
	"github.com/jetbrains-infra/packer-builder-vsphere/driver"
)

type HardwareConfig struct {
	CPUs           int32 `mapstructure:"CPUs"`
	CPUReservation int64 `mapstructure:"CPU_reservation"`
	CPULimit       int64 `mapstructure:"CPU_limit"`
	RAM            int64 `mapstructure:"RAM"`
	RAMReservation int64 `mapstructure:"RAM_reservation"`
	RAMReserveAll  bool  `mapstructure:"RAM_reserve_all"`
	DiskSize       int64 `mapstructure:"disk_size"`
}

func (c *HardwareConfig) Prepare() []error {
	var errs []error

	if c.RAMReservation > 0 && c.RAMReserveAll != false {
		errs = append(errs, fmt.Errorf("'RAM_reservation' and 'RAM_reserve_all' cannot be used together"))
	}

	return errs
}

type StepConfigureHardware struct {
	config *HardwareConfig
}

func (s *StepConfigureHardware) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	vm := state.Get("vm").(*driver.VirtualMachine)

	if *s.config != (HardwareConfig{}) {
		ui.Say("Customizing hardware parameters...")

		err := vm.Configure(&driver.HardwareConfig{
			CPUs:           s.config.CPUs,
			CPUReservation: s.config.CPUReservation,
			CPULimit:       s.config.CPULimit,
			RAM:            s.config.RAM,
			RAMReservation: s.config.RAMReservation,
			RAMReserveAll:  s.config.RAMReserveAll,
			DiskSize:       s.config.DiskSize,
		})
		if err != nil {
			state.Put("error", err)
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (s *StepConfigureHardware) Cleanup(multistep.StateBag) {}

package brkt

import (
	"fmt"

	"github.com/brkt/brkt-sdk-go/brkt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type stepGetMachineType struct {
	AvatarEnabled bool
	MinCpuCores   int
	MinRam        float64
	MachineType   string
}

func (s *stepGetMachineType) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	api := state.Get("api").(*brkt.API)

	ui.Say("Getting MachineType...")

	// machine type already set, go on
	if s.MachineType != "" {
		ui.Say("MachineType UUID supplied, using that MachineType...")

		if _, err := api.GetMachineType(s.MachineType); err != nil {
			state.Put("error", fmt.Errorf("could not find the supplied MachineType: %s", err))
			return multistep.ActionHalt
		}

		state.Put("machineType", s.MachineType)
		return multistep.ActionContinue
	}

	machineType, err := api.GetMachineTypeFromConstraint(&brkt.MachineTypeConstraint{
		MinRam:      s.MinRam,
		MinCpuCores: s.MinCpuCores,
		SupportsPV:  s.AvatarEnabled,
	})

	if err != nil {
		state.Put("error", fmt.Errorf("error while getting machine type matching CPU and RAM criteria: %s", err))
		return multistep.ActionHalt
	}
	if machineType == nil {
		state.Put("error", fmt.Errorf("no machine type found matching CPU and RAM criteria"))
		return multistep.ActionHalt
	}

	s.MachineType = machineType.Data.Id
	state.Put("machineType", s.MachineType)

	ui.Say(fmt.Sprintf("Selected instance type with %d cores and %.1f GB of RAM", machineType.Data.CpuCores, machineType.Data.Ram))

	if s.MachineType == "" {
		state.Put("error", fmt.Errorf("no machine type set"))
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepGetMachineType) Cleanup(state multistep.StateBag) {}

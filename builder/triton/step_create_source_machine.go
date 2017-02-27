package triton

import (
	"fmt"
	"time"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

// StepCreateSourceMachine creates an machine with the specified attributes
// and waits for it to become available for provisioners.
type StepCreateSourceMachine struct{}

func (s *StepCreateSourceMachine) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(Config)
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Creating source machine...")

	machineId, err := driver.CreateMachine(config)
	if err != nil {
		state.Put("error", fmt.Errorf("Problem creating source machine: %s", err))
		return multistep.ActionHalt
	}

	ui.Say("Waiting for source machine to become available...")
	err = driver.WaitForMachineState(machineId, "running", 10*time.Minute)
	if err != nil {
		state.Put("error", fmt.Errorf("Problem waiting for source machine to become available: %s", err))
		return multistep.ActionHalt
	}

	state.Put("machine", machineId)

	return multistep.ActionContinue
}

func (s *StepCreateSourceMachine) Cleanup(state multistep.StateBag) {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	machineIdRaw, ok := state.GetOk("machine")
	if ok && machineIdRaw.(string) != "" {
		machineId := machineIdRaw.(string)
		ui.Say(fmt.Sprintf("Stopping source machine (%s)...", machineId))
		err := driver.StopMachine(machineId)
		if err != nil {
			state.Put("error", fmt.Errorf("Problem stopping source machine: %s", err))
			return
		}

		ui.Say(fmt.Sprintf("Waiting for source machine to stop (%s)...", machineId))
		err = driver.WaitForMachineState(machineId, "stopped", 10*time.Minute)
		if err != nil {
			state.Put("error", fmt.Errorf("Problem waiting for source machine to stop: %s", err))
			return
		}

		ui.Say(fmt.Sprintf("Deleting source machine (%s)...", machineId))
		err = driver.DeleteMachine(machineId)
		if err != nil {
			state.Put("error", fmt.Errorf("Problem deleting source machine: %s", err))
			return
		}
	}
}

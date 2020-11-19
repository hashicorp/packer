package triton

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

// StepStopMachine stops the machine with the given Machine ID, and waits
// for it to reach the stopped state.
type StepStopMachine struct{}

func (s *StepStopMachine) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packersdk.Ui)

	machineId := state.Get("machine").(string)

	ui.Say(fmt.Sprintf("Stopping source machine (%s)...", machineId))
	err := driver.StopMachine(machineId)
	if err != nil {
		state.Put("error", fmt.Errorf("Problem stopping source machine: %s", err))
		return multistep.ActionHalt
	}

	ui.Say(fmt.Sprintf("Waiting for source machine to stop (%s)...", machineId))
	err = driver.WaitForMachineState(machineId, "stopped", 10*time.Minute)
	if err != nil {
		state.Put("error", fmt.Errorf("Problem waiting for source machine to stop: %s", err))
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepStopMachine) Cleanup(state multistep.StateBag) {
	// Explicitly don't clean up here as StepCreateSourceMachine will do it if necessary
	// and there is no real meaning to cleaning this up.
}

package triton

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

// StepDeleteMachine deletes the machine with the ID specified in state["machine"]
type StepDeleteMachine struct{}

func (s *StepDeleteMachine) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packersdk.Ui)

	machineId := state.Get("machine").(string)

	ui.Say("Deleting source machine...")
	err := driver.DeleteMachine(machineId)
	if err != nil {
		state.Put("error", fmt.Errorf("Problem deleting source machine: %s", err))
		return multistep.ActionHalt
	}

	ui.Say("Waiting for source machine to be deleted...")
	err = driver.WaitForMachineDeletion(machineId, 10*time.Minute)
	if err != nil {
		state.Put("error", fmt.Errorf("Problem waiting for source machine to be deleted: %s", err))
		return multistep.ActionHalt
	}

	state.Put("machine", "")

	return multistep.ActionContinue
}

func (s *StepDeleteMachine) Cleanup(state multistep.StateBag) {
	// No clean up to do here...
}

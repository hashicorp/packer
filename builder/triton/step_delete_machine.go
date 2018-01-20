package triton

import (
	"fmt"
	"time"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// StepDeleteMachine deletes the machine with the ID specified in state["machine"]
type StepDeleteMachine struct{}

func (s *StepDeleteMachine) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

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

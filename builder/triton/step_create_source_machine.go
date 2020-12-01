package triton

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

// StepCreateSourceMachine creates an machine with the specified attributes
// and waits for it to become available for provisioners.
type StepCreateSourceMachine struct{}

func (s *StepCreateSourceMachine) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packersdk.Ui)

	if !config.MachineImageFilters.Empty() {
		ui.Say("Selecting an image based on search criteria")
		imageId, err := driver.GetImage(*config)
		if err != nil {
			state.Put("error", fmt.Errorf("Problem selecting an image based on an search criteria: %s", err))
			return multistep.ActionHalt
		}
		ui.Say(fmt.Sprintf("Based, on given search criteria, Machine ID is: %q", imageId))
		config.MachineImage = imageId
	}

	machineId, err := driver.CreateMachine(*config)
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
	// instance_id is the generic term used so that users can have access to the
	// instance id inside of the provisioners, used in step_provision.
	state.Put("instance_id", machineId)
	return multistep.ActionContinue
}

func (s *StepCreateSourceMachine) Cleanup(state multistep.StateBag) {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packersdk.Ui)

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

		ui.Say(fmt.Sprintf("Waiting for source machine to be destroyed (%s)...", machineId))
		err = driver.WaitForMachineDeletion(machineId, 10*time.Minute)
		if err != nil {
			state.Put("error", fmt.Errorf("Problem waiting for source machine to be deleted: %s", err))
			return
		}
	}
}

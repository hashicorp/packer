package common

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type StepUnmountDvdDrive struct {
}

func (s *StepUnmountDvdDrive) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	vmName := state.Get("vmName").(string)
	ui := state.Get("ui").(packersdk.Ui)

	ui.Say("Unmount/delete os dvd drive...")

	dvdControllerState := state.Get("os.dvd.properties")

	if dvdControllerState == nil {
		return multistep.ActionContinue
	}

	dvdController := dvdControllerState.(DvdControllerProperties)

	if dvdController.Existing {
		ui.Say(fmt.Sprintf("Unmounting os dvd drives controller %d location %d ...",
			dvdController.ControllerNumber, dvdController.ControllerLocation))
		err := driver.UnmountDvdDrive(vmName, dvdController.ControllerNumber, dvdController.ControllerLocation)
		if err != nil {
			err := fmt.Errorf("Error unmounting os dvd drive: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	} else {
		ui.Say(fmt.Sprintf("Delete os dvd drives controller %d location %d ...",
			dvdController.ControllerNumber, dvdController.ControllerLocation))
		err := driver.DeleteDvdDrive(vmName, dvdController.ControllerNumber, dvdController.ControllerLocation)
		if err != nil {
			err := fmt.Errorf("Error deleting os dvd drive: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	state.Put("os.dvd.properties", nil)

	return multistep.ActionContinue
}

func (s *StepUnmountDvdDrive) Cleanup(state multistep.StateBag) {
}

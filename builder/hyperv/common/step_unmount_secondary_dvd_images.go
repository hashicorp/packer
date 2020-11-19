package common

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type StepUnmountSecondaryDvdImages struct {
}

func (s *StepUnmountSecondaryDvdImages) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packersdk.Ui)
	vmName := state.Get("vmName").(string)

	ui.Say("Unmount/delete secondary dvd drives...")

	dvdControllersState := state.Get("secondary.dvd.properties")

	if dvdControllersState == nil {
		return multistep.ActionContinue
	}

	dvdControllers := dvdControllersState.([]DvdControllerProperties)

	for _, dvdController := range dvdControllers {
		if dvdController.Existing {
			ui.Say(fmt.Sprintf("Unmounting secondary dvd drives controller %d location %d ...",
				dvdController.ControllerNumber, dvdController.ControllerLocation))
			err := driver.UnmountDvdDrive(vmName, dvdController.ControllerNumber, dvdController.ControllerLocation)
			if err != nil {
				err := fmt.Errorf("Error unmounting secondary dvd drive: %s", err)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}
		} else {
			ui.Say(fmt.Sprintf("Delete secondary dvd drives controller %d location %d ...",
				dvdController.ControllerNumber, dvdController.ControllerLocation))
			err := driver.DeleteDvdDrive(vmName, dvdController.ControllerNumber, dvdController.ControllerLocation)
			if err != nil {
				err := fmt.Errorf("Error deleting secondary dvd drive: %s", err)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}
		}
	}

	state.Put("secondary.dvd.properties", nil)

	return multistep.ActionContinue
}

func (s *StepUnmountSecondaryDvdImages) Cleanup(state multistep.StateBag) {
}

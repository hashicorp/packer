package common

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type StepUnmountGuestAdditions struct {
}

func (s *StepUnmountGuestAdditions) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	vmName := state.Get("vmName").(string)
	ui := state.Get("ui").(packersdk.Ui)

	ui.Say("Unmount/delete Integration Services dvd drive...")

	dvdControllerState := state.Get("guest.dvd.properties")

	if dvdControllerState == nil {
		return multistep.ActionContinue
	}

	dvdController := dvdControllerState.(DvdControllerProperties)

	if dvdController.Existing {
		ui.Say(fmt.Sprintf("Unmounting Integration Services dvd drives controller %d location %d ...",
			dvdController.ControllerNumber, dvdController.ControllerLocation))
		err := driver.UnmountDvdDrive(vmName, dvdController.ControllerNumber, dvdController.ControllerLocation)
		if err != nil {
			err := fmt.Errorf("Error unmounting Integration Services dvd drive: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	} else {
		ui.Say(fmt.Sprintf("Delete Integration Services dvd drives controller %d location %d ...",
			dvdController.ControllerNumber, dvdController.ControllerLocation))
		err := driver.DeleteDvdDrive(vmName, dvdController.ControllerNumber, dvdController.ControllerLocation)
		if err != nil {
			err := fmt.Errorf("Error deleting Integration Services dvd drive: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	state.Put("guest.dvd.properties", nil)

	return multistep.ActionContinue
}

func (s *StepUnmountGuestAdditions) Cleanup(state multistep.StateBag) {
}

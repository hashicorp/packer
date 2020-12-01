package common

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type StepMountGuestAdditions struct {
	GuestAdditionsMode string
	GuestAdditionsPath string
	Generation         uint
}

func (s *StepMountGuestAdditions) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)

	if s.GuestAdditionsMode != "attach" {
		ui.Say("Skipping mounting Integration Services Setup Disk...")
		return multistep.ActionContinue
	}

	driver := state.Get("driver").(Driver)
	ui.Say("Mounting Integration Services Setup Disk...")

	vmName := state.Get("vmName").(string)

	// should be able to mount up to 60 additional iso images using SCSI
	// but Windows would only allow a max of 22 due to available drive letters
	// Will Windows assign DVD drives to A: and B: ?

	// For IDE, there are only 2 controllers (0,1) with 2 locations each (0,1)

	var dvdControllerProperties DvdControllerProperties

	controllerNumber, controllerLocation, err := driver.CreateDvdDrive(vmName, s.GuestAdditionsPath, s.Generation)
	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	dvdControllerProperties.ControllerNumber = controllerNumber
	dvdControllerProperties.ControllerLocation = controllerLocation
	dvdControllerProperties.Existing = false
	state.Put("guest.dvd.properties", dvdControllerProperties)

	ui.Say(fmt.Sprintf("Mounting Integration Services dvd drive %s ...", s.GuestAdditionsPath))
	err = driver.MountDvdDrive(vmName, s.GuestAdditionsPath, controllerNumber, controllerLocation)
	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	log.Println(fmt.Sprintf("ISO %s mounted on DVD controller %v, location %v", s.GuestAdditionsPath,
		controllerNumber, controllerLocation))

	return multistep.ActionContinue
}

func (s *StepMountGuestAdditions) Cleanup(state multistep.StateBag) {
	if s.GuestAdditionsMode != "attach" {
		return
	}

	dvdControllerState := state.Get("guest.dvd.properties")

	if dvdControllerState == nil {
		return
	}

	dvdController := dvdControllerState.(DvdControllerProperties)
	ui := state.Get("ui").(packersdk.Ui)
	driver := state.Get("driver").(Driver)
	vmName := state.Get("vmName").(string)
	errorMsg := "Error unmounting Integration Services dvd drive: %s"

	ui.Say("Cleanup Integration Services dvd drive...")

	if dvdController.Existing {
		err := driver.UnmountDvdDrive(vmName, dvdController.ControllerNumber, dvdController.ControllerLocation)
		if err != nil {
			log.Print(fmt.Sprintf(errorMsg, err))
		}
	} else {
		err := driver.DeleteDvdDrive(vmName, dvdController.ControllerNumber, dvdController.ControllerLocation)
		if err != nil {
			log.Print(fmt.Sprintf(errorMsg, err))
		}
	}
}

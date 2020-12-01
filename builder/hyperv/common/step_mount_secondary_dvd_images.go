package common

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type StepMountSecondaryDvdImages struct {
	IsoPaths   []string
	Generation uint
}

type DvdControllerProperties struct {
	ControllerNumber   uint
	ControllerLocation uint
	Existing           bool
}

func (s *StepMountSecondaryDvdImages) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packersdk.Ui)
	ui.Say("Mounting secondary DVD images...")

	vmName := state.Get("vmName").(string)

	// should be able to mount up to 60 additional iso images using SCSI
	// but Windows would only allow a max of 22 due to available drive letters
	// Will Windows assign DVD drives to A: and B: ?

	// For IDE, there are only 2 controllers (0,1) with 2 locations each (0,1)
	var dvdProperties []DvdControllerProperties

	isoPaths := s.IsoPaths

	// Add our custom CD, if it exists
	cd_path, ok := state.Get("cd_path").(string)
	if ok {
		if cd_path != "" {
			isoPaths = append(isoPaths, cd_path)
		}
	}

	for _, isoPath := range isoPaths {
		var properties DvdControllerProperties

		controllerNumber, controllerLocation, err := driver.CreateDvdDrive(vmName, isoPath, s.Generation)
		if err != nil {
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		properties.ControllerNumber = controllerNumber
		properties.ControllerLocation = controllerLocation
		properties.Existing = false
		dvdProperties = append(dvdProperties, properties)
		state.Put("secondary.dvd.properties", dvdProperties)

		ui.Say(fmt.Sprintf("Mounting secondary dvd drive %s ...", isoPath))
		err = driver.MountDvdDrive(vmName, isoPath, controllerNumber, controllerLocation)
		if err != nil {
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		log.Println(fmt.Sprintf("ISO %s mounted on DVD controller %v, location %v", isoPath, controllerNumber,
			controllerLocation))
	}

	return multistep.ActionContinue
}

func (s *StepMountSecondaryDvdImages) Cleanup(state multistep.StateBag) {
	dvdControllersState := state.Get("secondary.dvd.properties")

	if dvdControllersState == nil {
		return
	}

	dvdControllers := dvdControllersState.([]DvdControllerProperties)
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packersdk.Ui)
	vmName := state.Get("vmName").(string)
	errorMsg := "Error unmounting secondary dvd drive: %s"

	ui.Say("Clean up secondary dvd drives...")

	for _, dvdController := range dvdControllers {

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
}

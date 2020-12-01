package common

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type StepMountDvdDrive struct {
	Generation      uint
	FirstBootDevice string
}

func (s *StepMountDvdDrive) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packersdk.Ui)

	errorMsg := "Error mounting dvd drive: %s"
	vmName := state.Get("vmName").(string)

	// Determine if we even have a dvd disk to attach
	var isoPath string
	if isoPathRaw, ok := state.GetOk("iso_path"); ok {
		isoPath = isoPathRaw.(string)
	} else {
		log.Println("No dvd disk, not attaching.")
		return multistep.ActionContinue
	}

	// Determine if its a virtual hdd to mount
	if strings.ToLower(filepath.Ext(isoPath)) == ".vhd" || strings.ToLower(filepath.Ext(isoPath)) == ".vhdx" {
		log.Println("Its a hard disk, not attaching.")
		return multistep.ActionContinue
	}

	// should be able to mount up to 60 additional iso images using SCSI
	// but Windows would only allow a max of 22 due to available drive letters
	// Will Windows assign DVD drives to A: and B: ?

	// For IDE, there are only 2 controllers (0,1) with 2 locations each (0,1)

	var dvdControllerProperties DvdControllerProperties
	controllerNumber, controllerLocation, err := driver.CreateDvdDrive(vmName, isoPath, s.Generation)
	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	dvdControllerProperties.ControllerNumber = controllerNumber
	dvdControllerProperties.ControllerLocation = controllerLocation
	dvdControllerProperties.Existing = false

	state.Put("os.dvd.properties", dvdControllerProperties)

	// the "first_boot_device" setting has precedence over the legacy boot order
	// configuration, but only if its been assigned a value.

	if s.FirstBootDevice == "" {

		if s.Generation > 1 {
			// only print this message for Gen2, it's not a true statement for Gen1 VMs
			ui.Say(fmt.Sprintf("Setting boot drive to os dvd drive %s ...", isoPath))
		}

		err = driver.SetBootDvdDrive(vmName, controllerNumber, controllerLocation, s.Generation)
		if err != nil {
			err := fmt.Errorf(errorMsg, err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

	}

	ui.Say(fmt.Sprintf("Mounting os dvd drive %s ...", isoPath))
	err = driver.MountDvdDrive(vmName, isoPath, controllerNumber, controllerLocation)
	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepMountDvdDrive) Cleanup(state multistep.StateBag) {
	dvdControllerState := state.Get("os.dvd.properties")

	if dvdControllerState == nil {
		return
	}

	dvdController := dvdControllerState.(DvdControllerProperties)
	driver := state.Get("driver").(Driver)
	vmName := state.Get("vmName").(string)
	ui := state.Get("ui").(packersdk.Ui)
	errorMsg := "Error unmounting os dvd drive: %s"

	ui.Say("Clean up os dvd drive...")

	if dvdController.Existing {
		err := driver.UnmountDvdDrive(vmName, dvdController.ControllerNumber, dvdController.ControllerLocation)
		if err != nil {
			err := fmt.Errorf("Error unmounting dvd drive: %s", err)
			log.Print(fmt.Sprintf(errorMsg, err))
		}
	} else {
		err := driver.DeleteDvdDrive(vmName, dvdController.ControllerNumber, dvdController.ControllerLocation)
		if err != nil {
			err := fmt.Errorf("Error deleting dvd drive: %s", err)
			log.Print(fmt.Sprintf(errorMsg, err))
		}
	}
}

// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package common

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
)

type StepMountDvdDrive struct {
	Generation    uint
	cleanup       bool
	dvdProperties DvdControllerProperties
}

func (s *StepMountDvdDrive) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	errorMsg := "Error mounting dvd drive: %s"
	vmName := state.Get("vmName").(string)
	isoPath := state.Get("iso_path").(string)

	// should be able to mount up to 60 additional iso images using SCSI
	// but Windows would only allow a max of 22 due to available drive letters
	// Will Windows assign DVD drives to A: and B: ?

	// For IDE, there are only 2 controllers (0,1) with 2 locations each (0,1)

	var dvdControllerProperties DvdControllerProperties
	controllerNumber, controllerLocation, err := driver.CreateDvdDrive(vmName, s.Generation)
	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	dvdControllerProperties.ControllerNumber = controllerNumber
	dvdControllerProperties.ControllerLocation = controllerLocation
	s.cleanup = true
	s.dvdProperties = dvdControllerProperties

	ui.Say(fmt.Sprintf("Setting boot drive to os dvd drive %s ..."), isoPath)
	err = driver.SetBootDvdDrive(vmName, controllerNumber, controllerLocation)
	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say(fmt.Sprintf("Mounting os dvd drive %s ...", isoPath))
	err = driver.MountDvdDrive(vmName, isoPath)
	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("os.dvd.properties", dvdControllerProperties)

	return multistep.ActionContinue
}

func (s *StepMountDvdDrive) Cleanup(state multistep.StateBag) {
	if !s.cleanup {
		return
	}

	driver := state.Get("driver").(Driver)

	errorMsg := "Error unmounting os dvd drive: %s"

	vmName := state.Get("vmName").(string)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Clean up os dvd drive...")

	dvdControllerProperties := s.dvdProperties

	if dvdControllerProperties.Existing {
		err := driver.UnmountDvdDrive(vmName)
		if err != nil {
			err := fmt.Errorf("Error unmounting dvd drive: %s", err)
			log.Print(fmt.Sprintf(errorMsg, err))
		}
	} else {
		err := driver.DeleteDvdDrive(vmName, dvdControllerProperties.ControllerNumber, dvdControllerProperties.ControllerLocation)
		if err != nil {
			err := fmt.Errorf("Error deleting dvd drive: %s", err)
			log.Print(fmt.Sprintf(errorMsg, err))
		}
	}
}

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

type StepMountSecondaryDvdImages struct {
	IsoPaths         []string
	Generation    uint
	cleanup bool
	dvdProperties []DvdControllerProperties
}

type DvdControllerProperties struct {
	ControllerNumber   uint
	ControllerLocation uint
	Existing bool
}

func (s *StepMountSecondaryDvdImages) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
	ui.Say("Mounting secondary DVD images...")

	vmName := state.Get("vmName").(string)

	// should be able to mount up to 60 additional iso images using SCSI
	// but Windows would only allow a max of 22 due to available drive letters
	// Will Windows assign DVD drives to A: and B: ?

	// For IDE, there are only 2 controllers (0,1) with 2 locations each (0,1)
	var dvdProperties []DvdControllerProperties

	for _, isoPath := range s.IsoPaths {
		var properties DvdControllerProperties

		controllerNumber, controllerLocation, err := driver.CreateDvdDrive(vmName, s.Generation)
		if err != nil {
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	
		properties.ControllerNumber = controllerNumber
		properties.ControllerLocation = controllerLocation
		
		s.cleanup = true
		dvdProperties = append(dvdProperties, properties)
		s.dvdProperties = dvdProperties		
	
		ui.Say(fmt.Sprintf("Mounting secondary dvd drive %s ...", isoPath))
		err = driver.MountDvdDriveByLocation(vmName, isoPath, controllerNumber, controllerLocation)
		if err != nil {
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	
		log.Println(fmt.Sprintf("ISO %s mounted on DVD controller %v, location %v", isoPath, controllerNumber, controllerLocation))
	}

	state.Put("secondary.dvd.properties", dvdProperties)
	
	return multistep.ActionContinue
}

func (s *StepMountSecondaryDvdImages) Cleanup(state multistep.StateBag) {
	if (!s.cleanup){
		return
	}
	
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
	ui.Say("Clean up secondary dvd drives...")

	vmName := state.Get("vmName").(string)

	errorMsg := "Error unmounting secondary dvd drive: %s"

	for _, dvdControllerProperties := range s.dvdProperties {
		
		if dvdControllerProperties.Existing {
			err := driver.UnmountDvdDrive(vmName)
			if err != nil {
				log.Print(fmt.Sprintf(errorMsg, err))
			}
		} else {
			err := driver.DeleteDvdDrive(vmName, dvdControllerProperties.ControllerNumber, dvdControllerProperties.ControllerLocation)
			if err != nil {
				log.Print(fmt.Sprintf(errorMsg, err))
			}
		}		
	}
}

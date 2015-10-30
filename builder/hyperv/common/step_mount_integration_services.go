// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package common

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	hyperv "github.com/mitchellh/packer/powershell/hyperv"
	"log"
	"os"
	"strconv"
)

type StepMountSecondaryDvdImages struct {
	Files         []string
	Generation    uint
	dvdProperties []DvdControllerProperties
}

type DvdControllerProperties struct {
	ControllerNumber   string
	ControllerLocation string
}

func (s *StepMountSecondaryDvdImages) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	ui.Say("Mounting secondary DVD images...")

	vmName := state.Get("vmName").(string)

	// should be able to mount up to 60 additional iso images using SCSI
	// but Windows would only allow a max of 22 due to available drive letters
	// Will Windows assign DVD drives to A: and B: ?

	// For IDE, there are only 2 controllers (0,1) with 2 locations each (0,1)
	dvdProperties, err := s.mountFiles(vmName)
	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	log.Println(fmt.Sprintf("Saving DVD properties %d DVDs", len(dvdProperties)))

	state.Put("secondary.dvd.properties", dvdProperties)

	return multistep.ActionContinue
}

func (s *StepMountSecondaryDvdImages) Cleanup(state multistep.StateBag) {

}

func (s *StepMountSecondaryDvdImages) mountFiles(vmName string) ([]DvdControllerProperties, error) {

	var dvdProperties []DvdControllerProperties

	properties, err := s.addAndMountIntegrationServicesSetupDisk(vmName)
	if err != nil {
		return dvdProperties, err
	}

	dvdProperties = append(dvdProperties, properties)

	for _, value := range s.Files {
		properties, err := s.addAndMountDvdDisk(vmName, value)
		if err != nil {
			return dvdProperties, err
		}

		dvdProperties = append(dvdProperties, properties)
	}

	return dvdProperties, nil
}

func (s *StepMountSecondaryDvdImages) addAndMountIntegrationServicesSetupDisk(vmName string) (DvdControllerProperties, error) {

	isoPath := os.Getenv("WINDIR") + "\\system32\\vmguest.iso"
	properties, err := s.addAndMountDvdDisk(vmName, isoPath)
	if err != nil {
		return properties, err
	}

	return properties, nil
}

func (s *StepMountSecondaryDvdImages) addAndMountDvdDisk(vmName string, isoPath string) (DvdControllerProperties, error) {
	var properties DvdControllerProperties

	controllerNumber, controllerLocation, err := hyperv.CreateDvdDrive(vmName, s.Generation)
	if err != nil {
		return properties, err
	}

	properties.ControllerNumber = strconv.FormatInt(int64(controllerNumber), 10)
	properties.ControllerLocation = strconv.FormatInt(int64(controllerLocation), 10)

	err = hyperv.MountDvdDriveByLocation(vmName, isoPath, controllerNumber, controllerLocation)
	if err != nil {
		return properties, err
	}

	log.Println(fmt.Sprintf("ISO %s mounted on DVD controller %v, location %v", isoPath, controllerNumber, controllerLocation))

	return properties, nil
}

// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package common

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type StepUnmountSecondaryDvdImages struct {
}

func (s *StepUnmountSecondaryDvdImages) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
	vmName := state.Get("vmName").(string)

	ui.Say("Unmounting secondary dvd drives...")

	dvdControllers := state.Get("secondary.dvd.properties").([]DvdControllerProperties)

	for _, dvdController := range dvdControllers {
		if dvdController.Existing {
			err := driver.UnmountDvdDrive(vmName)
			if err != nil {
				err := fmt.Errorf("Error unmounting secondary dvd drive: %s", err)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}
		} else {
			err := driver.DeleteDvdDrive(vmName, dvdController.ControllerNumber, dvdController.ControllerLocation)
			if err != nil {
				err := fmt.Errorf("Error deleting secondary dvd drive: %s", err)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}
		}		
	}

	return multistep.ActionContinue
}

func (s *StepUnmountSecondaryDvdImages) Cleanup(state multistep.StateBag) {
}

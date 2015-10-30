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

type StepUnmountGuestAdditions struct {
}

func (s *StepUnmountGuestAdditions) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	vmName := state.Get("vmName").(string)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Unmounting Integration Services dvd drive...")
		
	dvdController := state.Get("guest.dvd.properties").(DvdControllerProperties)
	
	if dvdController.Existing {
		err := driver.UnmountDvdDrive(vmName)
		if err != nil {
			err := fmt.Errorf("Error unmounting Integration Services dvd drive: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	} else {
		err := driver.DeleteDvdDrive(vmName, dvdController.ControllerNumber, dvdController.ControllerLocation)
		if err != nil {
			err := fmt.Errorf("Error deleting Integration Services dvd drive: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}
	
	return multistep.ActionContinue
}

func (s *StepUnmountGuestAdditions) Cleanup(state multistep.StateBag) {
}

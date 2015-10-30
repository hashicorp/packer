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

type StepUnmountSecondaryDvdImages struct {
}

func (s *StepUnmountSecondaryDvdImages) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
	ui.Say("Unmounting Integration Services Setup Disk...")

	vmName := state.Get("vmName").(string)

	// todo: should this message say removing the dvd?

	dvdProperties := state.Get("secondary.dvd.properties").([]DvdControllerProperties)

	log.Println(fmt.Sprintf("Found DVD properties %d", len(dvdProperties)))

	for _, dvdProperty := range dvdProperties {
		err := driver.DeleteDvdDrive(vmName, dvdProperty.ControllerNumber, dvdProperty.ControllerLocation)
		if err != nil {
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (s *StepUnmountSecondaryDvdImages) Cleanup(state multistep.StateBag) {
}

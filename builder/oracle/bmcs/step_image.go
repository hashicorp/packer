// Copyright (c) 2017 Oracle America, Inc.
// The contents of this file are subject to the Mozilla Public License Version
// 2.0 (the "License"); you may not use this file except in compliance with the
// License. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/

package bmcs

import (
	"fmt"

	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
)

type stepImage struct{}

func (s *stepImage) Run(state multistep.StateBag) multistep.StepAction {
	var (
		driver     = state.Get("driver").(Driver)
		ui         = state.Get("ui").(packer.Ui)
		instanceID = state.Get("instance_id").(string)
	)

	ui.Say("Creating image from instance...")

	image, err := driver.CreateImage(instanceID)
	if err != nil {
		state.Put("error", fmt.Errorf("Error creating image from instance: %s", err))
		return multistep.ActionHalt
	}

	err = driver.WaitForImageCreation(image.ID)
	if err != nil {
		state.Put("error", fmt.Errorf("Error waiting for image creation to finish: %s", err))
		return multistep.ActionHalt
	}

	// TODO(apryde): This is stale as .LifecycleState has changed to
	// AVAILABLE at this point. Does it matter?
	state.Put("image", image)

	ui.Say("Image created.")

	return multistep.ActionContinue
}

func (s *stepImage) Cleanup(state multistep.StateBag) {
	// Nothing to do
}

// Copyright (c) 2013 Kelsey Hightower. All rights reserved.
// Use of this source code is governed by the Apache License, Version 2.0
// that can be found in the LICENSE file.

package googlecompute

import (
	"fmt"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

// stepRegisterImage represents a Packer build step that registers GCE machine images.
type stepRegisterImage int

// Run executes the Packer build step that registers a GCE machine image.
func (s *stepRegisterImage) Run(state multistep.StateBag) multistep.StepAction {
	var (
		client = state.Get("client").(*GoogleComputeClient)
		config = state.Get("config").(config)
		ui     = state.Get("ui").(packer.Ui)
	)
	ui.Say("Adding image to the project...")
	imageURL := fmt.Sprintf("https://storage.cloud.google.com/%s/%s.tar.gz", config.BucketName, config.ImageName)
	operation, err := client.CreateImage(config.ImageName, config.ImageDescription, imageURL)
	if err != nil {
		err := fmt.Errorf("Error creating image: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	ui.Say("Waiting for image to become available...")
	err = waitForGlobalOperationState("DONE", operation.Name, client, config.stateTimeout)
	if err != nil {
		err := fmt.Errorf("Error creating image: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	state.Put("image_name", config.ImageName)
	return multistep.ActionContinue
}

// Cleanup.
func (s *stepRegisterImage) Cleanup(state multistep.StateBag) {}

// Copyright (c) 2013 Kelsey Hightower. All rights reserved.
// Use of this source code is governed by the Apache License, Version 2.0
// that can be found in the LICENSE file.

package googlecompute

import (
	"fmt"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

// stepUploadImage represents a Packer build step that uploads GCE machine images.
type stepUploadImage int

// Run executes the Packer build step that uploads a GCE machine image.
func (s *stepUploadImage) Run(state multistep.StateBag) multistep.StepAction {
	var (
		config        = state.Get("config").(config)
		comm          = state.Get("communicator").(packer.Communicator)
		sudoPrefix    = ""
		ui            = state.Get("ui").(packer.Ui)
		imageFilename = state.Get("image_file_name").(string)
	)
	ui.Say("Uploading image...")
	if config.SSHUsername != "root" {
		sudoPrefix = "sudo "
	}
	cmd := new(packer.RemoteCmd)
	cmd.Command = fmt.Sprintf("%s/usr/local/bin/gsutil cp %s gs://%s",
		sudoPrefix, imageFilename, config.BucketName)
	err := cmd.StartWithUi(comm, ui)
	if err != nil {
		err := fmt.Errorf("Error uploading image: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	return multistep.ActionContinue
}

// Cleanup.
func (s *stepUploadImage) Cleanup(state multistep.StateBag) {}

// Copyright (c) 2013 Kelsey Hightower. All rights reserved.
// Use of this source code is governed by the Apache License, Version 2.0
// that can be found in the LICENSE file.

package googlecompute

import (
	"fmt"
	"path/filepath"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

// stepCreateImage represents a Packer build step that creates GCE machine
// images.
type stepCreateImage int

// Run executes the Packer build step that creates a GCE machine image.
//
// Currently the only way to create a GCE image is to run the gcimagebundle
// command on the running GCE instance.
func (s *stepCreateImage) Run(state multistep.StateBag) multistep.StepAction {
	var (
		config     = state.Get("config").(config)
		comm       = state.Get("communicator").(packer.Communicator)
		sudoPrefix = ""
		ui         = state.Get("ui").(packer.Ui)
	)
	ui.Say("Creating image...")
	if config.SSHUsername != "root" {
		sudoPrefix = "sudo "
	}
	imageFilename := fmt.Sprintf("%s.tar.gz", config.ImageName)
	imageBundleCmd := "/usr/bin/gcimagebundle -d /dev/sda -o /tmp/"
	cmd := new(packer.RemoteCmd)
	cmd.Command = fmt.Sprintf("%s%s --output_file_name %s",
		sudoPrefix, imageBundleCmd, imageFilename)
	err := cmd.StartWithUi(comm, ui)
	if err != nil {
		err := fmt.Errorf("Error creating image: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	state.Put("image_file_name", filepath.Join("/tmp", imageFilename))
	return multistep.ActionContinue
}

func (s *stepCreateImage) Cleanup(state multistep.StateBag) {}

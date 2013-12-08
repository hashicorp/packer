// Copyright (c) 2013 Kelsey Hightower. All rights reserved.
// Use of this source code is governed by the Apache License, Version 2.0
// that can be found in the LICENSE file.

package googlecompute

import (
	"fmt"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

// stepUpdateGsutil represents a Packer build step that updates the gsutil
// utility to the latest version available.
type stepUpdateGsutil int

// Run executes the Packer build step that updates the gsutil utility to the
// latest version available.
//
// This step is required to prevent the image creation process from hanging;
// the image creation process utilizes the gcimagebundle cli tool which will
// prompt to update gsutil if a newer version is available.
func (s *stepUpdateGsutil) Run(state multistep.StateBag) multistep.StepAction {
	var (
		config     = state.Get("config").(config)
		comm       = state.Get("communicator").(packer.Communicator)
		sudoPrefix = ""
		ui         = state.Get("ui").(packer.Ui)
	)
	ui.Say("Updating gsutil...")
	if config.SSHUsername != "root" {
		sudoPrefix = "sudo "
	}
	gsutilUpdateCmd := "/usr/local/bin/gsutil update -n -f"
	cmd := new(packer.RemoteCmd)
	cmd.Command = fmt.Sprintf("%s%s", sudoPrefix, gsutilUpdateCmd)
	err := cmd.StartWithUi(comm, ui)
	if err != nil {
		err := fmt.Errorf("Error updating gsutil: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	return multistep.ActionContinue
}

// Cleanup.
func (s *stepUpdateGsutil) Cleanup(state multistep.StateBag) {}

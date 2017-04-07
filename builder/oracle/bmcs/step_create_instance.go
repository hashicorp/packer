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

type stepCreateInstance struct{}

func (s *stepCreateInstance) Run(state multistep.StateBag) multistep.StepAction {
	var (
		driver    = state.Get("driver").(Driver)
		ui        = state.Get("ui").(packer.Ui)
		publicKey = state.Get("publicKey").(string)
	)

	ui.Say("Creating instance...")

	instanceID, err := driver.CreateInstance(publicKey)
	if err != nil {
		state.Put("error", fmt.Errorf("Problem creating instance: %s", err))
		return multistep.ActionHalt
	}

	state.Put("instance_id", instanceID)

	ui.Say(fmt.Sprintf("Created instance (%s).", instanceID))

	ui.Say("Waiting for instance to enter 'RUNNING' state...")

	err = driver.WaitForInstanceState(instanceID, []string{"STARTING", "PROVISIONING"}, "RUNNING")
	if err != nil {
		state.Put("error", fmt.Errorf("Error waiting for instance to start:  %s", err))
		return multistep.ActionHalt
	}

	ui.Say("Instance 'RUNNING'.")

	return multistep.ActionContinue
}

func (s *stepCreateInstance) Cleanup(state multistep.StateBag) {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	idRaw, ok := state.GetOk("instance_id")
	if !ok {
		return
	}
	id := idRaw.(string)

	ui.Say(fmt.Sprintf("Terminating instance (%s)...", id))

	if err := driver.TerminateInstance(id); err != nil {
		state.Put("error", fmt.Sprintf("Error terminating instance. Please terminate manually: %s", err))
		return
	}

	err := driver.WaitForInstanceState(id, []string{"TERMINATING"}, "TERMINATED")
	if err != nil {
		state.Put("error", fmt.Sprintf("Error terminating instance. Please terminate manually: %s", err))
		return
	}

	ui.Say("Terminated instance.")
}

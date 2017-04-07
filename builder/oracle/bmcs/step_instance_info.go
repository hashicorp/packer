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

type stepInstanceInfo struct{}

func (s *stepInstanceInfo) Run(state multistep.StateBag) multistep.StepAction {
	var (
		driver = state.Get("driver").(Driver)
		ui     = state.Get("ui").(packer.Ui)
		id     = state.Get("instance_id").(string)
	)

	ip, err := driver.GetInstanceIP(id)
	if err != nil {
		state.Put("error", fmt.Errorf("Error getting instance's public IP: %s", err))
		return multistep.ActionHalt
	}

	state.Put("instance_ip", ip)

	ui.Say(fmt.Sprintf("Instance has public IP: %s.", ip))

	return multistep.ActionContinue
}

func (s *stepInstanceInfo) Cleanup(state multistep.StateBag) {
	// no cleanup
}

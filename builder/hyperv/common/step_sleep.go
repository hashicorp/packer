// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package common

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"time"
)

type StepSleep struct {
	Minutes time.Duration
	ActionName string
}

func (s *StepSleep) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	if(len(s.ActionName)>0){
		ui.Say(s.ActionName + "! Waiting for "+ fmt.Sprintf("%v",uint(s.Minutes)) + " minutes to let the action to complete...")
	}
	time.Sleep(time.Minute*s.Minutes);

	return multistep.ActionContinue
}

func (s *StepSleep) Cleanup(state multistep.StateBag) {

}

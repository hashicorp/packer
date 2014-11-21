// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package common

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"regexp"
	"strings"
)

type StepAcceptEula struct {
}

func (s *StepAcceptEula) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	errorMsg := "EULA agreement step error: %s"

	ans, err := ui.Ask("<TODO:validate text with Snesha> Do you accept the EULA? (type 'Yes' if you accept): ")

	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	eulaAnswer := strings.TrimSpace(ans)

	if len(eulaAnswer) == 0 {
		err := fmt.Errorf("Your answer is empty and Packer has to exit.")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt

	}else {
		pattern := "^[Yy][Ee][Ss]$"
		value := eulaAnswer

		match, _ := regexp.MatchString(pattern, value)
		if !match {
			err := fmt.Errorf("You should accept the EULA or Packer has to exit.")
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (s *StepAcceptEula) Cleanup(state multistep.StateBag) {
	// do nothing
}

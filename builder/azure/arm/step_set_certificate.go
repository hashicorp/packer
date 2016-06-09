// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in builder/azure for license information.

package arm

import (
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/builder/azure/common/constants"
	"github.com/mitchellh/packer/packer"
)

type StepSetCertificate struct {
	config *Config
	say    func(message string)
	error  func(e error)
}

func NewStepSetCertificate(config *Config, ui packer.Ui) *StepSetCertificate {
	var step = &StepSetCertificate{
		config: config,
		say:    func(message string) { ui.Say(message) },
		error:  func(e error) { ui.Error(e.Error()) },
	}

	return step
}

func (s *StepSetCertificate) Run(state multistep.StateBag) multistep.StepAction {
	s.say("Setting the certificate's URL ...")

	var winRMCertificateUrl = state.Get(constants.ArmCertificateUrl).(string)
	s.config.tmpWinRMCertificateUrl = winRMCertificateUrl

	return multistep.ActionContinue
}

func (*StepSetCertificate) Cleanup(multistep.StateBag) {
}

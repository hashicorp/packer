// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in builder/azure for license information.

package arm

import (
	"fmt"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/builder/azure/common/constants"
	"github.com/mitchellh/packer/packer"
)

type StepValidateTemplate struct {
	client   *AzureClient
	validate func(resourceGroupName string, deploymentName string, templateParameters *TemplateParameters) error
	say      func(message string)
	error    func(e error)
}

func NewStepValidateTemplate(client *AzureClient, ui packer.Ui) *StepValidateTemplate {
	var step = &StepValidateTemplate{
		client: client,
		say:    func(message string) { ui.Say(message) },
		error:  func(e error) { ui.Error(e.Error()) },
	}

	step.validate = step.validateTemplate
	return step
}

func (s *StepValidateTemplate) validateTemplate(resourceGroupName string, deploymentName string, templateParameters *TemplateParameters) error {
	factory := newDeploymentFactory(Linux)
	deployment, err := factory.create(*templateParameters)

	if err != nil {
		return err
	}

	_, err = s.client.Validate(resourceGroupName, deploymentName, *deployment)
	return err
}

func (s *StepValidateTemplate) Run(state multistep.StateBag) multistep.StepAction {
	s.say("Validating deployment template ...")

	var resourceGroupName = state.Get(constants.ArmResourceGroupName).(string)
	var deploymentName = state.Get(constants.ArmDeploymentName).(string)
	var templateParameters = state.Get(constants.ArmTemplateParameters).(*TemplateParameters)

	s.say(fmt.Sprintf(" -> ResourceGroupName : '%s'", resourceGroupName))
	s.say(fmt.Sprintf(" -> DeploymentName    : '%s'", deploymentName))

	err := s.validate(resourceGroupName, deploymentName, templateParameters)
	if err != nil {
		state.Put(constants.Error, err)
		s.error(err)

		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (*StepValidateTemplate) Cleanup(multistep.StateBag) {
}

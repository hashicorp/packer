// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in the project root for license information.

package arm

import (
	"fmt"

	"github.com/mitchellh/packer/builder/azure/common/constants"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type StepValidateTemplate struct {
	client   *AzureClient
	template string
	validate func(resourceGroupName string, deploymentName string, templateParameters *TemplateParameters) error
	say      func(message string)
	error    func(e error)
}

func NewStepValidateTemplate(client *AzureClient, ui packer.Ui, template string) *StepValidateTemplate {
	var step = &StepValidateTemplate{
		client:   client,
		template: template,
		say:      func(message string) { ui.Say(message) },
		error:    func(e error) { ui.Error(e.Error()) },
	}

	step.validate = step.validateTemplate
	return step
}

func (s *StepValidateTemplate) validateTemplate(resourceGroupName string, deploymentName string, templateParameters *TemplateParameters) error {
	factory := newDeploymentFactory(s.template)
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
	return processStepResult(err, s.error, state)
}

func (*StepValidateTemplate) Cleanup(multistep.StateBag) {
}

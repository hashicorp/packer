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
	validate func(resourceGroupName string, deploymentName string) error
	say      func(message string)
	error    func(e error)
	config   *Config
	factory  templateFactoryFunc
}

func NewStepValidateTemplate(client *AzureClient, ui packer.Ui, config *Config, factory templateFactoryFunc) *StepValidateTemplate {
	var step = &StepValidateTemplate{
		client:  client,
		say:     func(message string) { ui.Say(message) },
		error:   func(e error) { ui.Error(e.Error()) },
		config:  config,
		factory: factory,
	}

	step.validate = step.validateTemplate
	return step
}

func (s *StepValidateTemplate) validateTemplate(resourceGroupName string, deploymentName string) error {
	deployment, err := s.factory(s.config)
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

	s.say(fmt.Sprintf(" -> ResourceGroupName : '%s'", resourceGroupName))
	s.say(fmt.Sprintf(" -> DeploymentName    : '%s'", deploymentName))

	err := s.validate(resourceGroupName, deploymentName)
	return processStepResult(err, s.error, state)
}

func (*StepValidateTemplate) Cleanup(multistep.StateBag) {
}

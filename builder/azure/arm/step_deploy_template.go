// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in builder/azure for license information.

package arm

import (
	"fmt"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/builder/azure/common"
	"github.com/mitchellh/packer/builder/azure/common/constants"
	"github.com/mitchellh/packer/packer"
)

type StepDeployTemplate struct {
	client  *AzureClient
	deploy  func(resourceGroupName string, deploymentName string, cancelCh <-chan struct{}) error
	say     func(message string)
	error   func(e error)
	config  *Config
	factory templateFactoryFunc
}

func NewStepDeployTemplate(client *AzureClient, ui packer.Ui, config *Config, factory templateFactoryFunc) *StepDeployTemplate {
	var step = &StepDeployTemplate{
		client:  client,
		say:     func(message string) { ui.Say(message) },
		error:   func(e error) { ui.Error(e.Error()) },
		config:  config,
		factory: factory,
	}

	step.deploy = step.deployTemplate
	return step
}

func (s *StepDeployTemplate) deployTemplate(resourceGroupName string, deploymentName string, cancelCh <-chan struct{}) error {
	deployment, err := s.factory(s.config)
	if err != nil {
		return err
	}

	_, err = s.client.DeploymentsClient.CreateOrUpdate(resourceGroupName, deploymentName, *deployment, cancelCh)
	return err
}

func (s *StepDeployTemplate) Run(state multistep.StateBag) multistep.StepAction {
	s.say("Deploying deployment template ...")

	var resourceGroupName = state.Get(constants.ArmResourceGroupName).(string)
	var deploymentName = state.Get(constants.ArmDeploymentName).(string)

	s.say(fmt.Sprintf(" -> ResourceGroupName : '%s'", resourceGroupName))
	s.say(fmt.Sprintf(" -> DeploymentName    : '%s'", deploymentName))

	result := common.StartInterruptibleTask(
		func() bool { return common.IsStateCancelled(state) },
		func(cancelCh <-chan struct{}) error {
			return s.deploy(resourceGroupName, deploymentName, cancelCh)
		},
	)

	return processInterruptibleResult(result, s.error, state)
}

func (*StepDeployTemplate) Cleanup(multistep.StateBag) {
}

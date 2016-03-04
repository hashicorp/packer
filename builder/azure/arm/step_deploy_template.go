// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in the project root for license information.

package arm

import (
	"fmt"

	"github.com/mitchellh/packer/builder/azure/common/constants"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type StepDeployTemplate struct {
	client *AzureClient
	deploy func(resourceGroupName string, deploymentName string, templateParameters *TemplateParameters) error
	say    func(message string)
	error  func(e error)
}

func NewStepDeployTemplate(client *AzureClient, ui packer.Ui) *StepDeployTemplate {
	var step = &StepDeployTemplate{
		client: client,
		say:    func(message string) { ui.Say(message) },
		error:  func(e error) { ui.Error(e.Error()) },
	}

	step.deploy = step.deployTemplate
	return step
}

func (s *StepDeployTemplate) deployTemplate(resourceGroupName string, deploymentName string, templateParameters *TemplateParameters) error {
	factory := newDeploymentFactory(Linux)
	deployment, err := factory.create(*templateParameters)
	if err != nil {
		return err
	}

	res, err := s.client.DeploymentsClient.CreateOrUpdate(resourceGroupName, deploymentName, *deployment)
	if err != nil {
		return err
	}

	s.client.DeploymentsClient.PollAsNeeded(res.Response.Response)
	poller := NewDeploymentPoller(func() (string, error) {
		r, e := s.client.DeploymentsClient.Get(resourceGroupName, deploymentName)
		if r.Properties != nil && r.Properties.ProvisioningState != nil {
			return *r.Properties.ProvisioningState, e
		}

		return "UNKNOWN", e
	})

	pollStatus, err := poller.PollAsNeeded()
	if err != nil {
		return err
	}

	if pollStatus != DeploySucceeded {
		return fmt.Errorf("Deployment failed with a status of '%s'.", pollStatus)
	}

	return nil
}

func (s *StepDeployTemplate) Run(state multistep.StateBag) multistep.StepAction {
	s.say("Deploying deployment template ...")

	var resourceGroupName = state.Get(constants.ArmResourceGroupName).(string)
	var deploymentName = state.Get(constants.ArmDeploymentName).(string)
	var templateParameters = state.Get(constants.ArmTemplateParameters).(*TemplateParameters)

	s.say(fmt.Sprintf(" -> ResourceGroupName : '%s'", resourceGroupName))
	s.say(fmt.Sprintf(" -> DeploymentName    : '%s'", deploymentName))

	err := s.deploy(resourceGroupName, deploymentName, templateParameters)
	if err != nil {
		state.Put(constants.Error, err)
		s.error(err)

		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (*StepDeployTemplate) Cleanup(multistep.StateBag) {
}

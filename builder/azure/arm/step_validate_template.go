package arm

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/builder/azure/common/constants"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type StepValidateTemplate struct {
	client   *AzureClient
	validate func(ctx context.Context, resourceGroupName string, deploymentName string) error
	say      func(message string)
	error    func(e error)
	config   *Config
	factory  templateFactoryFunc
}

func NewStepValidateTemplate(client *AzureClient, ui packersdk.Ui, config *Config, factory templateFactoryFunc) *StepValidateTemplate {
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

func (s *StepValidateTemplate) validateTemplate(ctx context.Context, resourceGroupName string, deploymentName string) error {
	deployment, err := s.factory(s.config)
	if err != nil {
		return err
	}

	_, err = s.client.DeploymentsClient.Validate(ctx, resourceGroupName, deploymentName, *deployment)
	if err != nil {
		s.say(s.client.LastError.Error())
	}
	return err
}

func (s *StepValidateTemplate) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	s.say("Validating deployment template ...")

	var resourceGroupName = state.Get(constants.ArmResourceGroupName).(string)
	var deploymentName = state.Get(constants.ArmDeploymentName).(string)

	s.say(fmt.Sprintf(" -> ResourceGroupName : '%s'", resourceGroupName))
	s.say(fmt.Sprintf(" -> DeploymentName    : '%s'", deploymentName))

	err := s.validate(ctx, resourceGroupName, deploymentName)
	return processStepResult(err, s.error, state)
}

func (*StepValidateTemplate) Cleanup(multistep.StateBag) {
}

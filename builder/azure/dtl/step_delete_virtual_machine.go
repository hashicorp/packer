package dtl

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/builder/azure/common/constants"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type StepDeleteVirtualMachine struct {
	client *AzureClient
	config *Config
	delete func(ctx context.Context, resourceGroupName string, computeName string, state multistep.StateBag) error
	say    func(message string)
	error  func(e error)
}

func NewStepDeleteVirtualMachine(client *AzureClient, ui packersdk.Ui, config *Config) *StepDeleteVirtualMachine {
	var step = &StepDeleteVirtualMachine{
		client: client,
		config: config,
		say:    func(message string) { ui.Say(message) },
		error:  func(e error) { ui.Error(e.Error()) },
	}

	step.delete = step.deleteVirtualMachine
	return step
}

func (s *StepDeleteVirtualMachine) deleteVirtualMachine(ctx context.Context, resourceGroupName string, vmName string, state multistep.StateBag) error {
	f, err := s.client.DtlVirtualMachineClient.Delete(ctx, resourceGroupName, s.config.LabName, vmName)
	if err == nil {
		err = f.WaitForCompletionRef(ctx, s.client.DtlVirtualMachineClient.Client)
	}
	if err != nil {
		s.say("Error from delete VM")
		s.say(s.client.LastError.Error())
	}

	return err
}

func (s *StepDeleteVirtualMachine) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	s.say("Deleting the virtual machine ...")

	var resourceGroupName = state.Get(constants.ArmResourceGroupName).(string)
	var computeName = state.Get(constants.ArmComputeName).(string)

	s.say(fmt.Sprintf(" -> ResourceGroupName : '%s'", resourceGroupName))
	s.say(fmt.Sprintf(" -> ComputeName       : '%s'", computeName))

	err := s.deleteVirtualMachine(ctx, resourceGroupName, computeName, state)

	s.say("Deleting virtual machine ...Complete")
	return processStepResult(err, s.error, state)
}

func (*StepDeleteVirtualMachine) Cleanup(multistep.StateBag) {
}

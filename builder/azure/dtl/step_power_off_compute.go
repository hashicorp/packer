package dtl

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/builder/azure/common/constants"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type StepPowerOffCompute struct {
	client   *AzureClient
	config   *Config
	powerOff func(ctx context.Context, resourceGroupName string, labName, computeName string) error
	say      func(message string)
	error    func(e error)
}

func NewStepPowerOffCompute(client *AzureClient, ui packersdk.Ui, config *Config) *StepPowerOffCompute {

	var step = &StepPowerOffCompute{
		client: client,
		config: config,
		say:    func(message string) { ui.Say(message) },
		error:  func(e error) { ui.Error(e.Error()) },
	}

	step.powerOff = step.powerOffCompute
	return step
}

func (s *StepPowerOffCompute) powerOffCompute(ctx context.Context, resourceGroupName string, labName, computeName string) error {
	//f, err := s.client.VirtualMachinesClient.Deallocate(ctx, resourceGroupName, computeName)
	f, err := s.client.DtlVirtualMachineClient.Stop(ctx, resourceGroupName, labName, computeName)
	if err == nil {
		err = f.WaitForCompletionRef(ctx, s.client.DtlVirtualMachineClient.Client)
	}
	if err != nil {
		s.say(s.client.LastError.Error())
	}
	return err
}

func (s *StepPowerOffCompute) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	s.say("Powering off machine ...")

	var resourceGroupName = state.Get(constants.ArmResourceGroupName).(string)
	var computeName = state.Get(constants.ArmComputeName).(string)

	s.say(fmt.Sprintf(" -> ResourceGroupName : '%s'", resourceGroupName))
	s.say(fmt.Sprintf(" -> ComputeName       : '%s'", computeName))

	err := s.powerOff(ctx, s.config.LabResourceGroupName, s.config.LabName, computeName)

	s.say("Powering off machine ...Complete")
	return processStepResult(err, s.error, state)
}

func (*StepPowerOffCompute) Cleanup(multistep.StateBag) {
}

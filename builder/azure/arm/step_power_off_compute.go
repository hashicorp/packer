package arm

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/builder/azure/common/constants"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type StepPowerOffCompute struct {
	client   *AzureClient
	powerOff func(ctx context.Context, resourceGroupName string, computeName string) error
	say      func(message string)
	error    func(e error)
}

func NewStepPowerOffCompute(client *AzureClient, ui packer.Ui) *StepPowerOffCompute {
	var step = &StepPowerOffCompute{
		client: client,
		say:    func(message string) { ui.Say(message) },
		error:  func(e error) { ui.Error(e.Error()) },
	}

	step.powerOff = step.powerOffCompute
	return step
}

func (s *StepPowerOffCompute) powerOffCompute(ctx context.Context, resourceGroupName string, computeName string) error {
	f, err := s.client.VirtualMachinesClient.PowerOff(ctx, resourceGroupName, computeName)
	if err == nil {
		err = f.WaitForCompletion(ctx, s.client.VirtualMachinesClient.Client)
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

	err := s.powerOff(ctx, resourceGroupName, computeName)

	return processStepResult(err, s.error, state)
}

func (*StepPowerOffCompute) Cleanup(multistep.StateBag) {
}

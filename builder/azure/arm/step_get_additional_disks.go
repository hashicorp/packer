package arm

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2018-04-01/compute"

	"github.com/hashicorp/packer/builder/azure/common/constants"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type StepGetDataDisk struct {
	client *AzureClient
	query  func(ctx context.Context, resourceGroupName string, computeName string) (compute.VirtualMachine, error)
	say    func(message string)
	error  func(e error)
}

func NewStepGetAdditionalDisks(client *AzureClient, ui packersdk.Ui) *StepGetDataDisk {
	var step = &StepGetDataDisk{
		client: client,
		say:    func(message string) { ui.Say(message) },
		error:  func(e error) { ui.Error(e.Error()) },
	}

	step.query = step.queryCompute
	return step
}

func (s *StepGetDataDisk) queryCompute(ctx context.Context, resourceGroupName string, computeName string) (compute.VirtualMachine, error) {
	vm, err := s.client.VirtualMachinesClient.Get(ctx, resourceGroupName, computeName, "")
	if err != nil {
		s.say(s.client.LastError.Error())
	}
	return vm, err
}

func (s *StepGetDataDisk) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	s.say("Querying the machine's additional disks properties ...")

	var resourceGroupName = state.Get(constants.ArmResourceGroupName).(string)
	var computeName = state.Get(constants.ArmComputeName).(string)

	s.say(fmt.Sprintf(" -> ResourceGroupName : '%s'", resourceGroupName))
	s.say(fmt.Sprintf(" -> ComputeName       : '%s'", computeName))

	vm, err := s.query(ctx, resourceGroupName, computeName)
	if err != nil {
		state.Put(constants.Error, err)
		s.error(err)

		return multistep.ActionHalt
	}

	if vm.StorageProfile.DataDisks != nil {
		var vhdUri string
		additional_disks := make([]string, len(*vm.StorageProfile.DataDisks))
		for i, additionaldisk := range *vm.StorageProfile.DataDisks {
			if additionaldisk.Vhd != nil {
				vhdUri = *additionaldisk.Vhd.URI
				s.say(fmt.Sprintf(" -> Additional Disk %d          : '%s'", i+1, vhdUri))
			} else {
				vhdUri = *additionaldisk.ManagedDisk.ID
				s.say(fmt.Sprintf(" -> Managed Additional Disk %d  : '%s'", i+1, vhdUri))
			}
			additional_disks[i] = vhdUri
		}
		state.Put(constants.ArmAdditionalDiskVhds, additional_disks)
	}

	return multistep.ActionContinue
}

func (*StepGetDataDisk) Cleanup(multistep.StateBag) {
}

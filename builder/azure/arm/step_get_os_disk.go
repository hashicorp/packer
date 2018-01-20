package arm

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/arm/compute"

	"github.com/hashicorp/packer/builder/azure/common/constants"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type StepGetOSDisk struct {
	client *AzureClient
	query  func(resourceGroupName string, computeName string) (compute.VirtualMachine, error)
	say    func(message string)
	error  func(e error)
}

func NewStepGetOSDisk(client *AzureClient, ui packer.Ui) *StepGetOSDisk {
	var step = &StepGetOSDisk{
		client: client,
		say:    func(message string) { ui.Say(message) },
		error:  func(e error) { ui.Error(e.Error()) },
	}

	step.query = step.queryCompute
	return step
}

func (s *StepGetOSDisk) queryCompute(resourceGroupName string, computeName string) (compute.VirtualMachine, error) {
	vm, err := s.client.VirtualMachinesClient.Get(resourceGroupName, computeName, "")
	if err != nil {
		s.say(s.client.LastError.Error())
	}
	return vm, err
}

func (s *StepGetOSDisk) Run(state multistep.StateBag) multistep.StepAction {
	s.say("Querying the machine's properties ...")

	var resourceGroupName = state.Get(constants.ArmResourceGroupName).(string)
	var computeName = state.Get(constants.ArmComputeName).(string)

	s.say(fmt.Sprintf(" -> ResourceGroupName : '%s'", resourceGroupName))
	s.say(fmt.Sprintf(" -> ComputeName       : '%s'", computeName))

	vm, err := s.query(resourceGroupName, computeName)
	if err != nil {
		state.Put(constants.Error, err)
		s.error(err)

		return multistep.ActionHalt
	}

	var vhdUri string
	if vm.StorageProfile.OsDisk.Vhd != nil {
		vhdUri = *vm.StorageProfile.OsDisk.Vhd.URI
		s.say(fmt.Sprintf(" -> OS Disk           : '%s'", vhdUri))
	} else {
		vhdUri = *vm.StorageProfile.OsDisk.ManagedDisk.ID
		s.say(fmt.Sprintf(" -> Managed OS Disk   : '%s'", vhdUri))
	}

	state.Put(constants.ArmOSDiskVhd, vhdUri)
	return multistep.ActionContinue
}

func (*StepGetOSDisk) Cleanup(multistep.StateBag) {
}

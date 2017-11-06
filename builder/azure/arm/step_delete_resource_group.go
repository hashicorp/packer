package arm

import (
	"fmt"

	"github.com/Azure/go-autorest/autorest"
	"github.com/hashicorp/packer/builder/azure/common"
	"github.com/hashicorp/packer/builder/azure/common/constants"
	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
)

type StepDeleteResourceGroup struct {
	client *AzureClient
	delete func(state multistep.StateBag, resourceGroupName string, cancelCh <-chan struct{}) error
	say    func(message string)
	error  func(e error)
}

func NewStepDeleteResourceGroup(client *AzureClient, ui packer.Ui) *StepDeleteResourceGroup {
	var step = &StepDeleteResourceGroup{
		client: client,
		say:    func(message string) { ui.Say(message) },
		error:  func(e error) { ui.Error(e.Error()) },
	}

	step.delete = step.deleteResourceGroup
	return step
}

func (s *StepDeleteResourceGroup) deleteResourceGroup(state multistep.StateBag, resourceGroupName string, cancelCh <-chan struct{}) error {
	var err error
	if state.Get(constants.ArmIsExistingResourceGroup).(bool) {
		s.say("\nThe resource group was not created by Packer, only deleting individual resources ...")
		var deploymentName = state.Get(constants.ArmDeploymentName).(string)
		if deploymentName != "" {
			maxResources := int32(50)
			deploymentOperations, err := s.client.DeploymentOperationsClient.List(resourceGroupName, deploymentName, &maxResources)
			if err != nil {
				s.say(fmt.Sprintf("Error deleting resources.  Please delete them manually.\n\n"+
					"Name: %s\n"+
					"Error: %s", resourceGroupName, err))
				s.error(err)
			}
			for _, deploymentOperation := range *deploymentOperations.Value {
				// Sometimes an empty operation is added to the list by Azure
				if deploymentOperation.Properties.TargetResource == nil {
					continue
				}
				s.say(fmt.Sprintf(" -> %s : '%s'",
					*deploymentOperation.Properties.TargetResource.ResourceType,
					*deploymentOperation.Properties.TargetResource.ResourceName))
				var networkDeleteFunction func(string, string, <-chan struct{}) (<-chan autorest.Response, <-chan error)
				switch *deploymentOperation.Properties.TargetResource.ResourceType {
				case "Microsoft.Compute/virtualMachines":
					_, errChan := s.client.VirtualMachinesClient.Delete(resourceGroupName, *deploymentOperation.Properties.TargetResource.ResourceName, nil)
					err := <-errChan
					if err != nil {
						s.say(fmt.Sprintf("Error deleting resource.  Please delete manually.\n\n"+
							"Name: %s\n"+
							"Error: %s", *deploymentOperation.Properties.TargetResource.ResourceName, err.Error()))
						s.error(err)
					}
				case "Microsoft.Network/networkInterfaces":
					networkDeleteFunction = s.client.InterfacesClient.Delete
				case "Microsoft.Network/virtualNetworks":
					networkDeleteFunction = s.client.VirtualNetworksClient.Delete
				case "Microsoft.Network/publicIPAddresses":
					networkDeleteFunction = s.client.PublicIPAddressesClient.Delete
				}
				if networkDeleteFunction != nil {
					_, errChan := networkDeleteFunction(resourceGroupName, *deploymentOperation.Properties.TargetResource.ResourceName, nil)
					err := <-errChan
					if err != nil {
						s.say(fmt.Sprintf("Error deleting resource.  Please delete manually.\n\n"+
							"Name: %s\n"+
							"Error: %s", *deploymentOperation.Properties.TargetResource.ResourceName, err.Error()))
						s.error(err)
					}
				}
			}
		}
		return err
	} else {
		_, errChan := s.client.GroupsClient.Delete(resourceGroupName, cancelCh)
		s.say(state.Get(constants.ArmIsExistingResourceGroup).(string))
		err = <-errChan

		if err != nil {
			s.say(s.client.LastError.Error())
		}
		return err
	}
}

func (s *StepDeleteResourceGroup) Run(state multistep.StateBag) multistep.StepAction {
	s.say("Deleting resource group ...")

	var resourceGroupName = state.Get(constants.ArmResourceGroupName).(string)
	s.say(fmt.Sprintf(" -> ResourceGroupName : '%s'", resourceGroupName))

	result := common.StartInterruptibleTask(
		func() bool { return common.IsStateCancelled(state) },
		func(cancelCh <-chan struct{}) error { return s.delete(state, resourceGroupName, cancelCh) })
	stepAction := processInterruptibleResult(result, s.error, state)
	state.Put(constants.ArmIsResourceGroupCreated, false)

	return stepAction
}

func (*StepDeleteResourceGroup) Cleanup(multistep.StateBag) {
}

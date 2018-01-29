package arm

import (
	"context"
	"fmt"

	"github.com/Azure/go-autorest/autorest"
	"github.com/hashicorp/packer/builder/azure/common"
	"github.com/hashicorp/packer/builder/azure/common/constants"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

const (
	maxResourcesToDelete = 50
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
		err = s.deleteDeploymentResources(deploymentName, resourceGroupName)
		if err != nil {
			return err
		}

		if keyVaultDeploymentName, ok := state.GetOk(constants.ArmKeyVaultDeploymentName); ok {
			err = s.deleteDeploymentResources(keyVaultDeploymentName.(string), resourceGroupName)
			if err != nil {
				return err
			}
		}

		return nil
	} else {
		s.say("\nThe resource group was created by Packer, deleting ...")
		_, errChan := s.client.GroupsClient.Delete(resourceGroupName, cancelCh)
		err = <-errChan

		if err != nil {
			s.say(s.client.LastError.Error())
		}
		return err
	}
}

func (s *StepDeleteResourceGroup) deleteDeploymentResources(deploymentName, resourceGroupName string) error {
	maxResources := int32(maxResourcesToDelete)

	deploymentOperations, err := s.client.DeploymentOperationsClient.List(resourceGroupName, deploymentName, &maxResources)
	if err != nil {
		s.reportIfError(err, resourceGroupName)
		return err
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
		resourceName := *deploymentOperation.Properties.TargetResource.ResourceName

		switch *deploymentOperation.Properties.TargetResource.ResourceType {
		case "Microsoft.Compute/virtualMachines":
			_, errChan := s.client.VirtualMachinesClient.Delete(resourceGroupName, resourceName, nil)
			err := <-errChan
			s.reportIfError(err, resourceName)
		case "Microsoft.KeyVault/vaults":
			_, err := s.client.VaultClientDelete.Delete(resourceGroupName, resourceName)
			s.reportIfError(err, resourceName)
		case "Microsoft.Network/networkInterfaces":
			networkDeleteFunction = s.client.InterfacesClient.Delete
		case "Microsoft.Network/virtualNetworks":
			networkDeleteFunction = s.client.VirtualNetworksClient.Delete
		case "Microsoft.Network/publicIPAddresses":
			networkDeleteFunction = s.client.PublicIPAddressesClient.Delete
		}
		if networkDeleteFunction != nil {
			_, errChan := networkDeleteFunction(resourceGroupName, resourceName, nil)
			err := <-errChan
			s.reportIfError(err, resourceName)
		}
	}

	return nil
}

func (s *StepDeleteResourceGroup) reportIfError(err error, resourceName string) {
	if err != nil {
		s.say(fmt.Sprintf("Error deleting resource. Please delete manually.\n\n"+
			"Name: %s\n"+
			"Error: %s", resourceName, err.Error()))
		s.error(err)
	}
}

func (s *StepDeleteResourceGroup) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
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

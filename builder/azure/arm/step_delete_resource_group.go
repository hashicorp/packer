package arm

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/packer/builder/azure/common/constants"
	"github.com/hashicorp/packer/common/retry"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

const (
	maxResourcesToDelete = 50
)

type StepDeleteResourceGroup struct {
	client *AzureClient
	delete func(ctx context.Context, state multistep.StateBag, resourceGroupName string) error
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

func (s *StepDeleteResourceGroup) deleteResourceGroup(ctx context.Context, state multistep.StateBag, resourceGroupName string) error {
	var err error
	if state.Get(constants.ArmIsExistingResourceGroup).(bool) {
		s.say("\nThe resource group was not created by Packer, only deleting individual resources ...")
		var deploymentName = state.Get(constants.ArmDeploymentName).(string)
		err = s.deleteDeploymentResources(ctx, deploymentName, resourceGroupName)
		if err != nil {
			return err
		}

		if keyVaultDeploymentName, ok := state.GetOk(constants.ArmKeyVaultDeploymentName); ok {
			err = s.deleteDeploymentResources(ctx, keyVaultDeploymentName.(string), resourceGroupName)
			if err != nil {
				return err
			}
		}

		return nil
	} else {
		s.say("\nThe resource group was created by Packer, deleting ...")
		f, err := s.client.GroupsClient.Delete(ctx, resourceGroupName)
		if err == nil {
			if state.Get(constants.ArmAsyncResourceGroupDelete).(bool) {
				// No need to wait for the complition for delete if request is Accepted
				s.say(fmt.Sprintf("\nResource Group is being deleted, not waiting for deletion due to config. Resource Group Name '%s'", resourceGroupName))
			} else {
				f.WaitForCompletion(ctx, s.client.GroupsClient.Client)
			}

		}

		if err != nil {
			s.say(s.client.LastError.Error())
		}
		return err
	}
}

func (s *StepDeleteResourceGroup) deleteDeploymentResources(ctx context.Context, deploymentName, resourceGroupName string) error {
	maxResources := int32(maxResourcesToDelete)

	deploymentOperations, err := s.client.DeploymentOperationsClient.ListComplete(ctx, resourceGroupName, deploymentName, &maxResources)
	if err != nil {
		s.reportIfError(err, resourceGroupName)
		return err
	}

	for deploymentOperations.NotDone() {
		deploymentOperation := deploymentOperations.Value()
		// Sometimes an empty operation is added to the list by Azure
		if deploymentOperation.Properties.TargetResource == nil {
			deploymentOperations.Next()
			continue
		}

		resourceName := *deploymentOperation.Properties.TargetResource.ResourceName
		resourceType := *deploymentOperation.Properties.TargetResource.ResourceType

		s.say(fmt.Sprintf(" -> %s : '%s'",
			resourceType,
			resourceName))

		err := retry.Config{
			Tries:      10,
			RetryDelay: (&retry.Backoff{InitialBackoff: 10 * time.Second, MaxBackoff: 600 * time.Second, Multiplier: 2}).Linear,
		}.Run(ctx, func(ctx context.Context) error {
			err := deleteResource(ctx, s.client,
				resourceType,
				resourceName,
				resourceGroupName)
			if err != nil {
				s.reportIfError(err, resourceName)
			}
			return err
		})

		if err = deploymentOperations.Next(); err != nil {
			return err
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

func (s *StepDeleteResourceGroup) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	s.say("Deleting resource group ...")

	var resourceGroupName = state.Get(constants.ArmResourceGroupName).(string)
	s.say(fmt.Sprintf(" -> ResourceGroupName : '%s'", resourceGroupName))

	err := s.delete(ctx, state, resourceGroupName)
	if err != nil {
		state.Put(constants.Error, err)
		s.error(err)

		return multistep.ActionHalt
	}

	state.Put(constants.ArmIsResourceGroupCreated, false)

	return multistep.ActionContinue
}

func (*StepDeleteResourceGroup) Cleanup(multistep.StateBag) {
}

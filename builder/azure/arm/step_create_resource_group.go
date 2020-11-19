package arm

import (
	"context"
	"errors"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2018-02-01/resources"
	"github.com/hashicorp/packer/builder/azure/common/constants"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type StepCreateResourceGroup struct {
	client *AzureClient
	create func(ctx context.Context, resourceGroupName string, location string, tags map[string]*string) error
	say    func(message string)
	error  func(e error)
	exists func(ctx context.Context, resourceGroupName string) (bool, error)
}

func NewStepCreateResourceGroup(client *AzureClient, ui packersdk.Ui) *StepCreateResourceGroup {
	var step = &StepCreateResourceGroup{
		client: client,
		say:    func(message string) { ui.Say(message) },
		error:  func(e error) { ui.Error(e.Error()) },
	}

	step.create = step.createResourceGroup
	step.exists = step.doesResourceGroupExist
	return step
}

func (s *StepCreateResourceGroup) createResourceGroup(ctx context.Context, resourceGroupName string, location string, tags map[string]*string) error {
	_, err := s.client.GroupsClient.CreateOrUpdate(ctx, resourceGroupName, resources.Group{
		Location: &location,
		Tags:     tags,
	})

	if err != nil {
		s.say(s.client.LastError.Error())
	}
	return err
}

func (s *StepCreateResourceGroup) doesResourceGroupExist(ctx context.Context, resourceGroupName string) (bool, error) {
	exists, err := s.client.GroupsClient.CheckExistence(ctx, resourceGroupName)
	if err != nil {
		s.say(s.client.LastError.Error())
	}

	return exists.Response.StatusCode != 404, err
}

func (s *StepCreateResourceGroup) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	var doubleResource, ok = state.GetOk(constants.ArmDoubleResourceGroupNameSet)
	if ok && doubleResource.(bool) {
		err := errors.New("You have filled in both temp_resource_group_name and build_resource_group_name. Please choose one.")
		return processStepResult(err, s.error, state)
	}

	var resourceGroupName = state.Get(constants.ArmResourceGroupName).(string)
	var location = state.Get(constants.ArmLocation).(string)
	tags, ok := state.Get(constants.ArmTags).(map[string]*string)
	if !ok {
		err := fmt.Errorf("failed to extract tags from state bag")
		state.Put(constants.Error, err)
		s.error(err)
		return multistep.ActionHalt
	}

	exists, err := s.exists(ctx, resourceGroupName)
	if err != nil {
		return processStepResult(err, s.error, state)
	}
	configThinksExists := state.Get(constants.ArmIsExistingResourceGroup).(bool)
	if exists != configThinksExists {
		if configThinksExists {
			err = errors.New("The resource group you want to use does not exist yet. Please use temp_resource_group_name to create a temporary resource group.")
		} else {
			err = errors.New("A resource group with that name already exists. Please use build_resource_group_name to use an existing resource group.")
		}
		return processStepResult(err, s.error, state)
	}

	// If the resource group exists, we may not have permissions to update it so we don't.
	if !exists {
		s.say("Creating resource group ...")

		s.say(fmt.Sprintf(" -> ResourceGroupName : '%s'", resourceGroupName))
		s.say(fmt.Sprintf(" -> Location          : '%s'", location))
		s.say(fmt.Sprintf(" -> Tags              :"))
		for k, v := range tags {
			s.say(fmt.Sprintf(" ->> %s : %s", k, *v))
		}
		err = s.create(ctx, resourceGroupName, location, tags)
		if err == nil {
			state.Put(constants.ArmIsResourceGroupCreated, true)
		}
	} else {
		s.say("Using existing resource group ...")
		s.say(fmt.Sprintf(" -> ResourceGroupName : '%s'", resourceGroupName))
		s.say(fmt.Sprintf(" -> Location          : '%s'", location))
		state.Put(constants.ArmIsResourceGroupCreated, true)
	}

	return processStepResult(err, s.error, state)
}

func (s *StepCreateResourceGroup) Cleanup(state multistep.StateBag) {
	isCreated, ok := state.GetOk(constants.ArmIsResourceGroupCreated)
	if !ok || !isCreated.(bool) {
		return
	}

	ui := state.Get("ui").(packersdk.Ui)
	if state.Get(constants.ArmIsExistingResourceGroup).(bool) {
		ui.Say("\nThe resource group was not created by Packer, not deleting ...")
		return
	}

	ctx := context.TODO()
	resourceGroupName := state.Get(constants.ArmResourceGroupName).(string)
	if exists, err := s.exists(ctx, resourceGroupName); !exists || err != nil {
		return
	}

	ui.Say("\nCleanup requested, deleting resource group ...")
	f, err := s.client.GroupsClient.Delete(ctx, resourceGroupName)
	if err == nil {
		if state.Get(constants.ArmAsyncResourceGroupDelete).(bool) {
			s.say(fmt.Sprintf("\n Not waiting for Resource Group delete as requested by user. Resource Group Name is %s", resourceGroupName))
		} else {
			err = f.WaitForCompletionRef(ctx, s.client.GroupsClient.Client)
		}
	}
	if err != nil {
		ui.Error(fmt.Sprintf("Error deleting resource group.  Please delete it manually.\n\n"+
			"Name: %s\n"+
			"Error: %s", resourceGroupName, err))
		return
	}
	if !state.Get(constants.ArmAsyncResourceGroupDelete).(bool) {
		ui.Say("Resource group has been deleted.")
	}
}

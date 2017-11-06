package arm

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/arm/resources/resources"
	"github.com/hashicorp/packer/builder/azure/common/constants"
	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
)

type StepCreateResourceGroup struct {
	client *AzureClient
	create func(resourceGroupName string, location string, tags *map[string]*string) error
	say    func(message string)
	error  func(e error)
	exists func(resourceGroupName string) (bool, error)
}

func NewStepCreateResourceGroup(client *AzureClient, ui packer.Ui) *StepCreateResourceGroup {
	var step = &StepCreateResourceGroup{
		client: client,
		say:    func(message string) { ui.Say(message) },
		error:  func(e error) { ui.Error(e.Error()) },
	}

	step.create = step.createResourceGroup
	step.exists = step.doesResourceGroupExist
	return step
}

func (s *StepCreateResourceGroup) createResourceGroup(resourceGroupName string, location string, tags *map[string]*string) error {
	_, err := s.client.GroupsClient.CreateOrUpdate(resourceGroupName, resources.Group{
		Location: &location,
		Tags:     tags,
	})

	if err != nil {
		s.say(s.client.LastError.Error())
	}
	return err
}

func (s *StepCreateResourceGroup) doesResourceGroupExist(resourceGroupName string) (bool, error) {
	exists, err := s.client.GroupsClient.CheckExistence(resourceGroupName)
	if err != nil {
		s.say(s.client.LastError.Error())
	}

	return exists.Response.StatusCode != 404, err
}

func (s *StepCreateResourceGroup) Run(state multistep.StateBag) multistep.StepAction {
	s.say("Creating resource group ...")

	var resourceGroupName = state.Get(constants.ArmResourceGroupName).(string)
	var location = state.Get(constants.ArmLocation).(string)
	var tags = state.Get(constants.ArmTags).(*map[string]*string)

	s.say(fmt.Sprintf(" -> ResourceGroupName : '%s'", resourceGroupName))
	s.say(fmt.Sprintf(" -> Location          : '%s'", location))
	s.say(fmt.Sprintf(" -> Tags              :"))
	for k, v := range *tags {
		s.say(fmt.Sprintf(" ->> %s : %s", k, *v))
	}

	exists, err := s.exists(resourceGroupName)
	if err != nil {
		s.say(s.client.LastError.Error())
	}
	state.Put(constants.ArmIsExistingResourceGroup, exists)
	// If the resource group exists, we may not have permissions to update it so we don't.
	if !exists {
		err = s.create(resourceGroupName, location, tags)
		if err == nil {
			state.Put(constants.ArmIsResourceGroupCreated, true)
		}
	} else {
		// Mark the resource group as created to deal with later checks
		state.Put(constants.ArmIsResourceGroupCreated, true)
	}

	return processStepResult(err, s.error, state)
}

func (s *StepCreateResourceGroup) Cleanup(state multistep.StateBag) {
	isCreated, ok := state.GetOk(constants.ArmIsResourceGroupCreated)
	if !ok || !isCreated.(bool) {
		return
	}

	ui := state.Get("ui").(packer.Ui)
	if state.Get(constants.ArmIsExistingResourceGroup).(bool) {
		ui.Say("\nThe resource group was not created by Packer, not deleting ...")
		return
	} else {
		ui.Say("\nCleanup requested, deleting resource group ...")

		var resourceGroupName = state.Get(constants.ArmResourceGroupName).(string)
		_, errChan := s.client.GroupsClient.Delete(resourceGroupName, nil)
		err := <-errChan

		if err != nil {
			ui.Error(fmt.Sprintf("Error deleting resource group.  Please delete it manually.\n\n"+
				"Name: %s\n"+
				"Error: %s", resourceGroupName, err))
		}

		ui.Say("Resource group has been deleted.")
	}
}

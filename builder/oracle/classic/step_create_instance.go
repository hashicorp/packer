package classic

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-oracle-terraform/compute"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepCreateInstance struct{}

func (s *stepCreateInstance) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	// get variables from state
	ui := state.Get("ui").(packer.Ui)
	ui.Say("Creating Instance...")

	config := state.Get("config").(*Config)
	client := state.Get("client").(*compute.ComputeClient)
	ipAddName := state.Get("ipres_name").(string)
	secListName := state.Get("security_list").(string)

	netInfo := compute.NetworkingInfo{
		Nat:      []string{ipAddName},
		SecLists: []string{secListName},
	}

	// get instances client
	instanceClient := client.Instances()

	// Instances Input
	input := &compute.CreateInstanceInput{
		Name:       config.ImageName,
		Shape:      config.Shape,
		ImageList:  config.SourceImageList,
		Networking: map[string]compute.NetworkingInfo{"eth0": netInfo},
		Attributes: config.attribs,
	}
	if config.Comm.Type == "ssh" {
		keyName := state.Get("key_name").(string)
		input.SSHKeys = []string{keyName}
	}

	instanceInfo, err := instanceClient.CreateInstance(input)
	if err != nil {
		err = fmt.Errorf("Problem creating instance: %s", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

	state.Put("instance_info", instanceInfo)
	state.Put("instance_id", instanceInfo.ID)
	ui.Message(fmt.Sprintf("Created instance: %s.", instanceInfo.ID))
	return multistep.ActionContinue
}

func (s *stepCreateInstance) Cleanup(state multistep.StateBag) {
	// terminate instance
	ui := state.Get("ui").(packer.Ui)
	client := state.Get("client").(*compute.ComputeClient)
	config := state.Get("config").(*Config)
	imID := state.Get("instance_id").(string)

	ui.Say("Terminating source instance...")

	instanceClient := client.Instances()
	input := &compute.DeleteInstanceInput{
		Name: config.ImageName,
		ID:   imID,
	}

	err := instanceClient.DeleteInstance(input)
	if err != nil {
		err = fmt.Errorf("Problem destroying instance: %s", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return
	}
	// TODO wait for instance state to change to deleted?
	ui.Say("Terminated instance.")
}

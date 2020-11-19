package classic

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-oracle-terraform/compute"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type stepCreateInstance struct{}

func (s *stepCreateInstance) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	// get variables from state
	ui := state.Get("ui").(packersdk.Ui)
	ui.Say("Creating Instance...")

	config := state.Get("config").(*Config)
	client := state.Get("client").(*compute.Client)
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
		Entry:      config.SourceImageListEntry,
		Networking: map[string]compute.NetworkingInfo{"eth0": netInfo},
		Attributes: config.attribs,
	}
	if config.Comm.Type == "ssh" {
		input.SSHKeys = []string{config.Comm.SSHKeyPairName}
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
	instanceID, ok := state.GetOk("instance_id")
	if !ok {
		return
	}

	// terminate instance
	ui := state.Get("ui").(packersdk.Ui)
	client := state.Get("client").(*compute.Client)
	config := state.Get("config").(*Config)

	ui.Say("Terminating source instance...")

	instanceClient := client.Instances()
	input := &compute.DeleteInstanceInput{
		Name: config.ImageName,
		ID:   instanceID.(string),
	}

	err := instanceClient.DeleteInstance(input)
	if err != nil {
		err = fmt.Errorf("Problem destroying instance: %s", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return
	}
	ui.Say("Terminated instance.")
}

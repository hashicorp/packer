package classic

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-oracle-terraform/compute"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepCreatePVBuilder struct {
	name              string
	masterVolumeName  string
	builderVolumeName string
}

func (s *stepCreatePVBuilder) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	// get variables from state
	ui := state.Get("ui").(packer.Ui)
	ui.Say("Creating builder instance...")

	config := state.Get("config").(*Config)
	client := state.Get("client").(*compute.ComputeClient)
	ipAddName := state.Get("ipres_name").(string)
	secListName := state.Get("security_list").(string)

	// get instances client
	instanceClient := client.Instances()

	// Instances Input
	input := &compute.CreateInstanceInput{
		Name:  s.name,
		Shape: config.Shape,
		Networking: map[string]compute.NetworkingInfo{
			"eth0": compute.NetworkingInfo{
				Nat:      []string{ipAddName},
				SecLists: []string{secListName},
			},
		},
		Storage: []compute.StorageAttachmentInput{
			{
				Volume: s.builderVolumeName,
				Index:  1,
			},
			{
				Volume: s.masterVolumeName,
				Index:  2,
			},
		},
		ImageList:  config.SourceImageList,
		Attributes: config.attribs,
		SSHKeys:    []string{config.Comm.SSHKeyPairName},
	}

	instanceInfo, err := instanceClient.CreateInstance(input)
	if err != nil {
		err = fmt.Errorf("Problem creating instance: %s", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

	state.Put("builder_instance_info", instanceInfo)
	state.Put("builder_instance_id", instanceInfo.ID)
	ui.Message(fmt.Sprintf("Created builder instance: %s.", instanceInfo.ID))
	return multistep.ActionContinue
}

func (s *stepCreatePVBuilder) Cleanup(state multistep.StateBag) {
	instanceID, ok := state.GetOk("builder_instance_id")
	if !ok {
		return
	}

	// terminate instance
	ui := state.Get("ui").(packer.Ui)
	client := state.Get("client").(*compute.ComputeClient)
	config := state.Get("config").(*Config)

	ui.Say("Terminating builder instance...")

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
	// TODO wait for instance state to change to deleted?
	ui.Say("Terminated builder instance.")
}

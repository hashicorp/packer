package classic

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-oracle-terraform/compute"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type stepCreatePVMaster struct {
	Name            string
	VolumeName      string
	SecurityListKey string
}

func (s *stepCreatePVMaster) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	ui.Say("Creating master instance...")

	config := state.Get("config").(*Config)
	client := state.Get("client").(*compute.Client)
	ipAddName := state.Get("ipres_name").(string)
	secListName := state.Get(s.SecurityListKey).(string)

	// get instances client
	instanceClient := client.Instances()

	// Instances Input
	input := &compute.CreateInstanceInput{
		Name:  s.Name,
		Shape: config.Shape,
		Networking: map[string]compute.NetworkingInfo{
			"eth0": {
				Nat:      []string{ipAddName},
				SecLists: []string{secListName},
			},
		},
		Storage: []compute.StorageAttachmentInput{
			{
				Volume: s.VolumeName,
				Index:  1,
			},
		},
		BootOrder:  []int{1},
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

	state.Put("master_instance_info", instanceInfo)
	state.Put("master_instance_id", instanceInfo.ID)
	ui.Message(fmt.Sprintf("Created master instance: %s.", instanceInfo.Name))
	return multistep.ActionContinue
}

func (s *stepCreatePVMaster) Cleanup(state multistep.StateBag) {
	if _, deleted := state.GetOk("master_instance_deleted"); deleted {
		return
	}

	instanceID, ok := state.GetOk("master_instance_id")
	if !ok {
		return
	}

	// terminate instance
	ui := state.Get("ui").(packersdk.Ui)
	client := state.Get("client").(*compute.Client)

	ui.Say("Terminating builder instance...")

	instanceClient := client.Instances()
	input := &compute.DeleteInstanceInput{
		Name: s.Name,
		ID:   instanceID.(string),
	}

	err := instanceClient.DeleteInstance(input)
	if err != nil {
		err = fmt.Errorf("Problem destroying instance: %s", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return
	}
	ui.Say("Terminated master instance.")
}

package classic

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-oracle-terraform/compute"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepCreatePVMaster struct {
	name       string
	volumeName string
}

func (s *stepCreatePVMaster) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	// get variables from state
	ui := state.Get("ui").(packer.Ui)
	ui.Say("Creating master instance...")

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
				Volume: s.volumeName,
				Index:  1,
			},
		},
		BootOrder:  []int{1},
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

	state.Put("master_instance_info", instanceInfo)
	state.Put("master_instance_id", instanceInfo.ID)
	ui.Message(fmt.Sprintf("Created master instance: %s.", instanceInfo.ID))
	return multistep.ActionContinue
}

func (s *stepCreatePVMaster) Cleanup(state multistep.StateBag) {
}

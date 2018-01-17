package classic

import (
	"fmt"

	"github.com/hashicorp/go-oracle-terraform/compute"
	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
)

type stepCreateInstance struct{}

func (s *stepCreateInstance) Run(state multistep.StateBag) multistep.StepAction {
	ui.Say("Creating Instance...")

	// get variables from state
	ui := state.Get("ui").(packer.Ui)
	config := state.Get("config").(Config)
	client := state.Get("client").(*compute.ComputeClient)
	sshPublicKey := state.Get("publicKey").(string)

	// get instances client
	instanceClient := client.Instances()

	// Instances Input
	input := &compute.CreateInstanceInput{
		Name:       config.ImageName,
		Shape:      config.Shape,
		ImageList:  config.ImageList,
		SSHKeys:    []string{},
		Attributes: map[string]interface{}{},
	}

	instanceInfo, err := instanceClient.CreateInstance(input)
	if err != nil {
		err = fmt.Errorf("Problem creating instance: %s", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

	state.Put("instance_id", instanceInfo.ID)

	ui.Say(fmt.Sprintf("Created instance (%s).", instanceID))
}

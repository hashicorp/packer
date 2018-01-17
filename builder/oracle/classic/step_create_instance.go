package classic

import (
	"fmt"

	"github.com/hashicorp/go-oracle-terraform/compute"
	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
)

type stepCreateInstance struct{}

func (s *stepCreateInstance) Run(state multistep.StateBag) multistep.StepAction {
	// get variables from state
	ui := state.Get("ui").(packer.Ui)
	ui.Say("Creating Instance...")
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
		SSHKeys:    []string{sshPublicKey},
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
	ui.Say(fmt.Sprintf("Created instance (%s).", instanceInfo.ID))
	return multistep.ActionContinue
}

func (s *stepCreateInstance) Cleanup(state multistep.StateBag) {
	// Nothing to do
}

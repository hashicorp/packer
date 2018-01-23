package classic

import (
	"fmt"
	"log"

	"github.com/hashicorp/go-oracle-terraform/compute"
	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
)

type stepCreateInstance struct{}

func (s *stepCreateInstance) Run(state multistep.StateBag) multistep.StepAction {
	// get variables from state
	ui := state.Get("ui").(packer.Ui)
	ui.Say("Creating Instance...")
	config := state.Get("config").(*Config)
	client := state.Get("client").(*compute.ComputeClient)
	keyName := state.Get("key_name").(string)

	ipAddName := fmt.Sprintf("ipres_%s", config.ImageName)
	// secListName := "Megan_packer_test" // hack to get working; fix b4 release
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
		ImageList:  config.ImageList,
		SSHKeys:    []string{keyName},
		Networking: map[string]compute.NetworkingInfo{"eth0": netInfo},
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
	// terminate instance
	ui := state.Get("ui").(packer.Ui)
	client := state.Get("client").(*compute.ComputeClient)
	config := state.Get("config").(*Config)
	imID := state.Get("instance_id").(string)

	ui.Say(fmt.Sprintf("Terminating instance (%s)...", imID))

	instanceClient := client.Instances()
	// Instances Input
	input := &compute.DeleteInstanceInput{
		Name: config.ImageName,
		ID:   imID,
	}
	log.Printf("instance destroy input is %#v", input)

	err := instanceClient.DeleteInstance(input)
	if err != nil {
		err = fmt.Errorf("Problem destroying instance: %s", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return
	}
	// TODO wait for instance state to change to deleted?
	ui.Say("Terminated instance.")
	return
}

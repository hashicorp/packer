package classic

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

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
	keyName := state.Get("key_name").(string)
	ipAddName := state.Get("ipres_name").(string)
	secListName := state.Get("security_list").(string)

	netInfo := compute.NetworkingInfo{
		Nat:      []string{ipAddName},
		SecLists: []string{secListName},
	}

	// get instances client
	instanceClient := client.Instances()

	var data map[string]interface{}

	if config.Attributes != "" {
		err := json.Unmarshal([]byte(config.Attributes), &data)
		if err != nil {
			err = fmt.Errorf("Problem parsing json from attributes: %s", err)
			ui.Error(err.Error())
			state.Put("error", err)
			return multistep.ActionHalt
		}
	} else if config.AttributesFile != "" {
		fidata, err := ioutil.ReadFile(config.AttributesFile)
		if err != nil {
			err = fmt.Errorf("Problem reading attributes_file: %s", err)
			ui.Error(err.Error())
			state.Put("error", err)
			return multistep.ActionHalt
		}
		err = json.Unmarshal(fidata, &data)
		if err != nil {
			err = fmt.Errorf("Problem parsing json from attrinutes_file: %s", err)
			ui.Error(err.Error())
			state.Put("error", err)
			return multistep.ActionHalt
		}
	}

	// Instances Input
	input := &compute.CreateInstanceInput{
		Name:       config.ImageName,
		Shape:      config.Shape,
		ImageList:  config.SourceImageList,
		SSHKeys:    []string{keyName},
		Networking: map[string]compute.NetworkingInfo{"eth0": netInfo},
		Attributes: data,
	}

	instanceInfo, err := instanceClient.CreateInstance(input)
	if err != nil {
		err = fmt.Errorf("Problem creating instance: %s", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

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

package classic

import (
	"fmt"
	"log"
	"strings"

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

	// SSH KEY CONFIGURATION

	// grab packer-generated key from statebag context.
	sshPublicKey := strings.TrimSpace(state.Get("publicKey").(string))

	// form API call to add key to compute cloud
	sshKeyName := fmt.Sprintf("/Compute-%s/%s/packer_generated_key", config.IdentityDomain, config.Username)

	sshKeysClient := client.SSHKeys()
	sshKeysInput := compute.CreateSSHKeyInput{
		Name:    sshKeyName,
		Key:     sshPublicKey,
		Enabled: true,
	}

	// Load the packer-generated SSH key into the Oracle Compute cloud.
	keyInfo, err := sshKeysClient.CreateSSHKey(&sshKeysInput)
	if err != nil {
		// Key already exists; update key instead of creating it
		if strings.Contains(err.Error(), "packer_generated_key already exists") {
			updateKeysInput := compute.UpdateSSHKeyInput{
				Name:    sshKeyName,
				Key:     sshPublicKey,
				Enabled: true,
			}
			keyInfo, err = sshKeysClient.UpdateSSHKey(&updateKeysInput)
		} else {
			err = fmt.Errorf("Problem adding Public SSH key through Oracle's API: %s", err)
			ui.Error(err.Error())
			state.Put("error", err)
			return multistep.ActionHalt
		}
	}

	// NETWORKING INFO CONFIGURATION
	ipAddName := fmt.Sprintf("ipres_%s", config.ImageName)
	log.Printf("MEGAN ipADDName is %s", ipAddName)
	secListName := "Megan_packer_test"

	netInfo := compute.NetworkingInfo{
		Nat:      []string{ipAddName},
		SecLists: []string{secListName},
	}
	fmt.Sprintf("Megan netInfo is %#v", netInfo)

	// INSTANCE LAUNCH

	// get instances client
	instanceClient := client.Instances()

	// Instances Input
	input := &compute.CreateInstanceInput{
		Name:       config.ImageName,
		Shape:      config.Shape,
		ImageList:  config.ImageList,
		SSHKeys:    []string{keyInfo.Name},
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
	imID := state.Get("instance_id").(string)

	ui.Say(fmt.Sprintf("Terminating instance (%s)...", id))

	instanceClient := client.Instances()
	// Instances Input
	input := &compute.DeleteInstanceInput{
		Name: config.ImageName,
		ID:   imID,
	}

	instanceInfo, err := instanceClient.DeleteInstance(input)
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

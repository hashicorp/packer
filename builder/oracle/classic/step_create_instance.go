package classic

import (
	"fmt"
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
	sshPublicKey := strings.TrimSpace(state.Get("publicKey").(string))

	// Load the dynamically-generated SSH key into the Oracle Compute cloud.
	sshKeyName := fmt.Sprintf("/Compute-%s/%s/packer_generated_key", config.IdentityDomain, config.Username)

	sshKeysClient := client.SSHKeys()
	sshKeysInput := compute.CreateSSHKeyInput{
		Name:    sshKeyName,
		Key:     sshPublicKey,
		Enabled: true,
	}
	keyInfo, err := sshKeysClient.CreateSSHKey(&sshKeysInput)
	if err != nil {
		// Key already exists; update key instead
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

	// get instances client
	instanceClient := client.Instances()

	// Instances Input
	input := &compute.CreateInstanceInput{
		Name:      config.ImageName,
		Shape:     config.Shape,
		ImageList: config.ImageList,
		SSHKeys:   []string{keyInfo.Name},
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

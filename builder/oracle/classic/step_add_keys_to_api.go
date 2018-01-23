package classic

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-oracle-terraform/compute"
	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
)

type stepAddKeysToAPI struct{}

func (s *stepAddKeysToAPI) Run(state multistep.StateBag) multistep.StepAction {
	// get variables from state
	ui := state.Get("ui").(packer.Ui)
	ui.Say("Adding SSH keys to API...")
	config := state.Get("config").(*Config)
	client := state.Get("client").(*compute.ComputeClient)

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
	state.Put("key_name", keyInfo.Name)
	return multistep.ActionContinue
}

func (s *stepAddKeysToAPI) Cleanup(state multistep.StateBag) {
	// Nothing to do
}

package classic

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/go-oracle-terraform/compute"
	uuid "github.com/hashicorp/go-uuid"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepAddKeysToAPI struct{}

func (s *stepAddKeysToAPI) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	// get variables from state
	ui := state.Get("ui").(packer.Ui)
	config := state.Get("config").(*Config)
	client := state.Get("client").(*compute.ComputeClient)

	// grab packer-generated key from statebag context.
	sshPublicKey := strings.TrimSpace(state.Get("publicKey").(string))

	// form API call to add key to compute cloud
	uuid_string, err := uuid.GenerateUUID()
	if err != nil {
		ui.Error(fmt.Sprintf("Error creating unique SSH key name: %s", err.Error()))
		return multistep.ActionHalt
	}
	sshKeyName := fmt.Sprintf("/Compute-%s/%s/packer_generated_key_%s",
		config.IdentityDomain, config.Username, uuid_string)

	ui.Say(fmt.Sprintf("Creating temporary key: %s", sshKeyName))

	sshKeysClient := client.SSHKeys()
	sshKeysInput := compute.CreateSSHKeyInput{
		Name:    sshKeyName,
		Key:     sshPublicKey,
		Enabled: true,
	}

	// Load the packer-generated SSH key into the Oracle Compute cloud.
	keyInfo, err := sshKeysClient.CreateSSHKey(&sshKeysInput)
	if err != nil {
		err = fmt.Errorf("Problem adding Public SSH key through Oracle's API: %s", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}
	state.Put("key_name", keyInfo.Name)
	return multistep.ActionContinue
}

func (s *stepAddKeysToAPI) Cleanup(state multistep.StateBag) {
	// Delete the keys we created during this run
	keyName := state.Get("key_name").(string)
	ui := state.Get("ui").(packer.Ui)
	ui.Say("Deleting SSH keys...")
	deleteInput := compute.DeleteSSHKeyInput{Name: keyName}
	client := state.Get("client").(*compute.ComputeClient)
	deleteClient := client.SSHKeys()
	err := deleteClient.DeleteSSHKey(&deleteInput)
	if err != nil {
		ui.Error(fmt.Sprintf("Error deleting SSH keys: %s", err.Error()))
	}
	return
}

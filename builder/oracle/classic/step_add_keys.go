package classic

import (
	"bytes"
	"context"
	"fmt"

	"github.com/hashicorp/go-oracle-terraform/compute"
	"github.com/hashicorp/packer/common/uuid"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepAddKeysToAPI struct{}

func (s *stepAddKeysToAPI) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	// get variables from state
	ui := state.Get("ui").(packer.Ui)
	config := state.Get("config").(*Config)
	client := state.Get("client").(*compute.ComputeClient)

	if config.Comm.Type != "ssh" {
		ui.Say("Not using SSH communicator; skip generating SSH keys...")
		return multistep.ActionContinue
	}

	// grab packer-generated key from statebag context.
	sshPublicKey := bytes.TrimSpace(config.Comm.SSHPublicKey)

	// form API call to add key to compute cloud
	sshKeyName := config.Identifier(fmt.Sprintf("packer_generated_key_%s", uuid.TimeOrderedUUID()))

	ui.Say(fmt.Sprintf("Creating temporary key: %s", sshKeyName))

	sshKeysClient := client.SSHKeys()
	sshKeysInput := compute.CreateSSHKeyInput{
		Name:    sshKeyName,
		Key:     string(sshPublicKey),
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
	config.Comm.SSHKeyPairName = keyInfo.Name
	return multistep.ActionContinue
}

func (s *stepAddKeysToAPI) Cleanup(state multistep.StateBag) {
	// Delete the keys we created during this run
	config := state.Get("config").(*Config)
	if len(config.Comm.SSHKeyPairName) == 0 {
		// No keys were generated; none need to be cleaned up.
		return
	}
	ui := state.Get("ui").(packer.Ui)
	ui.Say("Deleting SSH keys...")
	deleteInput := compute.DeleteSSHKeyInput{Name: config.Comm.SSHKeyPairName}
	client := state.Get("client").(*compute.ComputeClient)
	deleteClient := client.SSHKeys()
	err := deleteClient.DeleteSSHKey(&deleteInput)
	if err != nil {
		ui.Error(fmt.Sprintf("Error deleting SSH keys: %s", err.Error()))
	}
	return
}

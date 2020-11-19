package classic

import (
	"bytes"
	"context"
	"fmt"

	"github.com/hashicorp/go-oracle-terraform/compute"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type stepAddKeysToAPI struct {
	Skip    bool
	KeyName string
}

func (s *stepAddKeysToAPI) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	// get variables from state
	ui := state.Get("ui").(packersdk.Ui)
	config := state.Get("config").(*Config)
	client := state.Get("client").(*compute.Client)

	if s.Skip {
		ui.Say("Skipping generating SSH keys...")
		return multistep.ActionContinue
	}
	// grab packer-generated key from statebag context.
	// Always check configured communicator for keys
	sshPublicKey := bytes.TrimSpace(config.Comm.SSHPublicKey)

	// form API call to add key to compute cloud

	ui.Say(fmt.Sprintf("Creating temporary key: %s", s.KeyName))

	sshKeysClient := client.SSHKeys()
	sshKeysInput := compute.CreateSSHKeyInput{
		Name:    s.KeyName,
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
	if s.Skip {
		return
	}
	config := state.Get("config").(*Config)
	// Delete the keys we created during this run
	if len(config.Comm.SSHKeyPairName) == 0 {
		// No keys were generated; none need to be cleaned up.
		return
	}
	ui := state.Get("ui").(packersdk.Ui)
	ui.Say("Deleting SSH keys...")
	deleteInput := compute.DeleteSSHKeyInput{Name: config.Comm.SSHKeyPairName}
	client := state.Get("client").(*compute.Client)
	deleteClient := client.SSHKeys()
	err := deleteClient.DeleteSSHKey(&deleteInput)
	if err != nil {
		ui.Error(fmt.Sprintf("Error deleting SSH keys: %s", err.Error()))
	}
	return
}

package digitalocean

import (
	"context"
	"fmt"
	"log"

	"github.com/digitalocean/godo"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/uuid"
)

type stepCreateSSHKey struct {
	keyId int
}

func (s *stepCreateSSHKey) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*godo.Client)
	ui := state.Get("ui").(packersdk.Ui)
	c := state.Get("config").(*Config)

	if c.Comm.SSHPublicKey == nil {
		ui.Say("No public SSH key found; skipping SSH public key import...")
		return multistep.ActionContinue
	}

	ui.Say("Importing SSH public key...")

	// The name of the public key on DO
	name := fmt.Sprintf("packer-%s", uuid.TimeOrderedUUID())

	// Create the key!
	key, _, err := client.Keys.Create(context.TODO(), &godo.KeyCreateRequest{
		Name:      name,
		PublicKey: string(c.Comm.SSHPublicKey),
	})
	if err != nil {
		err := fmt.Errorf("Error creating temporary SSH key: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// We use this to check cleanup
	s.keyId = key.ID

	log.Printf("temporary ssh key name: %s", name)

	// Remember some state for the future
	state.Put("ssh_key_id", key.ID)

	return multistep.ActionContinue
}

func (s *stepCreateSSHKey) Cleanup(state multistep.StateBag) {
	// If no key name is set, then we never created it, so just return
	if s.keyId == 0 {
		return
	}

	client := state.Get("client").(*godo.Client)
	ui := state.Get("ui").(packersdk.Ui)

	ui.Say("Deleting temporary ssh key...")
	_, err := client.Keys.DeleteByID(context.TODO(), s.keyId)
	if err != nil {
		log.Printf("Error cleaning up ssh key: %s", err)
		ui.Error(fmt.Sprintf(
			"Error cleaning up ssh key. Please delete the key manually: %s", err))
	}
}

package digitalocean

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type stepCreateDroplet struct {
	dropletId uint
}

func (s *stepCreateDroplet) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*DigitalOceanClient)
	ui := state.Get("ui").(packer.Ui)
	c := state.Get("config").(config)
	sshKeyId := state.Get("ssh_key_id").(uint)

	ui.Say("Creating droplet...")

	// Create the droplet based on configuration
	dropletId, err := client.CreateDroplet(c.DropletName, c.Size, c.Image, c.Region, sshKeyId, c.PrivateNetworking)

	if err != nil {
		err := fmt.Errorf("Error creating droplet: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// We use this in cleanup
	s.dropletId = dropletId

	// Store the droplet id for later
	state.Put("droplet_id", dropletId)

	return multistep.ActionContinue
}

func (s *stepCreateDroplet) Cleanup(state multistep.StateBag) {
	// If the dropletid isn't there, we probably never created it
	if s.dropletId == 0 {
		return
	}

	client := state.Get("client").(*DigitalOceanClient)
	ui := state.Get("ui").(packer.Ui)
	c := state.Get("config").(config)

	// Destroy the droplet we just created
	ui.Say("Destroying droplet...")

	err := client.DestroyDroplet(s.dropletId)
	if err != nil {
		curlstr := fmt.Sprintf("curl '%v/droplets/%v/destroy?client_id=%v&api_key=%v'",
			DIGITALOCEAN_API_URL, s.dropletId, c.ClientID, c.APIKey)

		ui.Error(fmt.Sprintf(
			"Error destroying droplet. Please destroy it manually: %v", curlstr))
	}
}

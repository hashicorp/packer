package digitalocean

import (
	"fmt"

	"github.com/digitalocean/godo"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type stepCreateDroplet struct {
	dropletId int
}

func (s *stepCreateDroplet) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*godo.Client)
	ui := state.Get("ui").(packer.Ui)
	c := state.Get("config").(Config)
	sshKeyId := state.Get("ssh_key_id").(int)

	// Create the droplet based on configuration
	ui.Say("Creating droplet...")
	droplet, _, err := client.Droplets.Create(&godo.DropletCreateRequest{
		Name:   c.DropletName,
		Region: c.Region,
		Size:   c.Size,
		Image: godo.DropletCreateImage{
			Slug: c.Image,
		},
		SSHKeys: []godo.DropletCreateSSHKey{
			godo.DropletCreateSSHKey{ID: int(sshKeyId)},
		},
		PrivateNetworking: c.PrivateNetworking,
		UserData:          c.UserData,
	})
	if err != nil {
		err := fmt.Errorf("Error creating droplet: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// We use this in cleanup
	s.dropletId = droplet.ID

	// Store the droplet id for later
	state.Put("droplet_id", droplet.ID)

	return multistep.ActionContinue
}

func (s *stepCreateDroplet) Cleanup(state multistep.StateBag) {
	// If the dropletid isn't there, we probably never created it
	if s.dropletId == 0 {
		return
	}

	client := state.Get("client").(*godo.Client)
	ui := state.Get("ui").(packer.Ui)

	// Destroy the droplet we just created
	ui.Say("Destroying droplet...")
	_, err := client.Droplets.Delete(s.dropletId)
	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error destroying droplet. Please destroy it manually: %s", err))
	}
}

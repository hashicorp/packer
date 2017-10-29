package digitalocean

import (
	"context"
	"fmt"
	"time"

	"io/ioutil"

	"github.com/digitalocean/godo"
	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
)

type stepCreateDroplet struct {
	dropletId int
	volumeIds []string
}

func (s *stepCreateDroplet) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*godo.Client)
	ui := state.Get("ui").(packer.Ui)
	c := state.Get("config").(Config)
	sshKeyId := state.Get("ssh_key_id").(int)
	volumeIds := state.Get("volume_ids").([]string)

	// Create the droplet based on configuration
	ui.Say("Creating droplet...")

	userData := c.UserData
	if c.UserDataFile != "" {
		contents, err := ioutil.ReadFile(c.UserDataFile)
		if err != nil {
			state.Put("error", fmt.Errorf("Problem reading user data file: %s", err))
			return multistep.ActionHalt
		}

		userData = string(contents)
	}

	volumes := []godo.DropletCreateVolume{}
	for _, volumeId := range volumeIds {
		volumes = append(volumes, godo.DropletCreateVolume{ID: volumeId})
	}

	droplet, _, err := client.Droplets.Create(context.TODO(), &godo.DropletCreateRequest{
		Name:   c.DropletName,
		Region: c.Region,
		Size:   c.Size,
		Image: godo.DropletCreateImage{
			Slug: c.Image,
		},
		SSHKeys: []godo.DropletCreateSSHKey{
			{ID: sshKeyId},
		},
		PrivateNetworking: c.PrivateNetworking,
		Monitoring:        c.Monitoring,
		UserData:          userData,
		Volumes:           volumes,
	})
	if err != nil {
		err := fmt.Errorf("Error creating droplet: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// We use these in cleanup
	s.dropletId = droplet.ID
	s.volumeIds = droplet.VolumeIDs

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

	// Manually wait for any attached volumes to detach before destroying the
	// image. Destroying a volume that's attached to a droplet throws an error.
	// (Destroying a droplet before destroying its volumes would theoretically
	// work, but the DigitalOcean API provides no way to wait for a droplet to
	// be destroyed.)
	if len(s.volumeIds) > 0 {
		ui.Say("Detaching volumes...")
		for i, volumeId := range s.volumeIds {
			action, _, err := client.StorageActions.DetachByDropletID(context.TODO(), volumeId, s.dropletId)
			if err != nil {
				ui.Error(fmt.Sprintf(
					"Error detaching volume %d. You may need to destroy it manually: %s",
					i, err))
			}
			ui.Say(fmt.Sprintf("Waiting for volume %d to detach...", i))
			if err := waitForActionState(godo.ActionCompleted, s.dropletId, action.ID,
				client, time.Minute); err != nil {
				err := fmt.Errorf("Error waiting for volume %d to detach: %s", i, err)
				ui.Error(err.Error())
			}
		}
	}

	// Destroy the droplet we just created
	ui.Say("Destroying droplet...")
	_, err := client.Droplets.Delete(context.TODO(), s.dropletId)
	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error destroying droplet. Please destroy it manually: %s", err))
	}
}

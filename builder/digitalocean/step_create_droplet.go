package digitalocean

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"io/ioutil"

	"github.com/digitalocean/godo"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

type stepCreateDroplet struct {
	dropletId int
}

func (s *stepCreateDroplet) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*godo.Client)
	ui := state.Get("ui").(packersdk.Ui)
	c := state.Get("config").(*Config)
	sshKeyId := state.Get("ssh_key_id").(int)

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

	createImage := getImageType(c.Image)

	dropletCreateReq := &godo.DropletCreateRequest{
		Name:   c.DropletName,
		Region: c.Region,
		Size:   c.Size,
		Image:  createImage,
		SSHKeys: []godo.DropletCreateSSHKey{
			{ID: sshKeyId},
		},
		PrivateNetworking: c.PrivateNetworking,
		Monitoring:        c.Monitoring,
		IPv6:              c.IPv6,
		UserData:          userData,
		Tags:              c.Tags,
		VPCUUID:           c.VPCUUID,
	}

	log.Printf("[DEBUG] Droplet create paramaters: %s", godo.Stringify(dropletCreateReq))

	droplet, _, err := client.Droplets.Create(context.TODO(), dropletCreateReq)
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
	// instance_id is the generic term used so that users can have access to the
	// instance id inside of the provisioners, used in step_provision.
	state.Put("instance_id", droplet.ID)

	return multistep.ActionContinue
}

func (s *stepCreateDroplet) Cleanup(state multistep.StateBag) {
	// If the dropletid isn't there, we probably never created it
	if s.dropletId == 0 {
		return
	}

	client := state.Get("client").(*godo.Client)
	ui := state.Get("ui").(packersdk.Ui)

	// Destroy the droplet we just created
	ui.Say("Destroying droplet...")
	_, err := client.Droplets.Delete(context.TODO(), s.dropletId)
	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error destroying droplet. Please destroy it manually: %s", err))
	}
}

func getImageType(image string) godo.DropletCreateImage {
	createImage := godo.DropletCreateImage{Slug: image}

	imageId, err := strconv.Atoi(image)
	if err == nil {
		createImage = godo.DropletCreateImage{ID: imageId}
	}

	return createImage
}

package openstack

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"github.com/rackspace/gophercloud"
	"log"
	"time"
)

type stepCreateImage struct{}

func (s *stepCreateImage) Run(state map[string]interface{}) multistep.StepAction {
	accessor := state["accessor"].(*gophercloud.Access)
	api := state["api"].(*gophercloud.ApiCriteria)
	config := state["config"].(config)
	server := state["server"].(*gophercloud.Server)
	ui := state["ui"].(packer.Ui)

	// Create the image
	ui.Say(fmt.Sprintf("Creating the image: %s", config.ImageName))
	createOpts := gophercloud.CreateImage{
		Name: config.ImageName,
	}
	csp, err := gophercloud.ServersApi(accessor, *api)
	imageId, err := csp.CreateImage(server.Id, createOpts)
	if err != nil {
		err := fmt.Errorf("Error creating image: %s", err)
		state["error"] = err
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Set the Image ID in the state
	ui.Say(fmt.Sprintf("Image: %s", imageId))
	state["image"] = imageId

	// Wait for the image to become ready
	ui.Say("Waiting for image to become ready...")
	if err := WaitForImage(accessor, api, imageId); err != nil {
		err := fmt.Errorf("Error waiting for image: %s", err)
		state["error"] = err
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepCreateImage) Cleanup(map[string]interface{}) {
	// No cleanup...
}

// WaitForImage waits for the given Image ID to become ready.
func WaitForImage(accessor *gophercloud.Access, api *gophercloud.ApiCriteria, imageId string) error {
	for {
		csp, err := gophercloud.ServersApi(accessor, *api)
		image, err := csp.ImageById(imageId)
		if err != nil {
			return err
		}

		if image.Status == "ACTIVE" {
			return nil
		}

		log.Printf("Waiting for image creation status: %s (%d%%)", image.Status, image.Progress)
		time.Sleep(2 * time.Second)
	}
}

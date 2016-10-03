package openstack

import (
	"fmt"
	"log"
	"time"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/openstack/compute/v2/images"
	"github.com/rackspace/gophercloud/openstack/compute/v2/servers"
)

type stepCreateImage struct{}

func (s *stepCreateImage) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(Config)
	server := state.Get("server").(*servers.Server)
	ui := state.Get("ui").(packer.Ui)

	// We need the v2 compute client
	client, err := config.computeV2Client()
	if err != nil {
		err = fmt.Errorf("Error initializing compute client: %s", err)
		state.Put("error", err)
		return multistep.ActionHalt
	}

	// Create the image
	ui.Say(fmt.Sprintf("Creating the image: %s", config.ImageName))
	imageId, err := servers.CreateImage(client, server.ID, servers.CreateImageOpts{
		Name:     config.ImageName,
		Metadata: config.ImageMetadata,
	}).ExtractImageID()
	if err != nil {
		err := fmt.Errorf("Error creating image: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Set the Image ID in the state
	ui.Message(fmt.Sprintf("Image: %s", imageId))
	state.Put("image", imageId)

	// Wait for the image to become ready
	ui.Say(fmt.Sprintf("Waiting for image %s (image id: %s) to become ready...", config.ImageName, imageId))
	if err := WaitForImage(client, imageId); err != nil {
		err := fmt.Errorf("Error waiting for image: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepCreateImage) Cleanup(multistep.StateBag) {
	// No cleanup...
}

// WaitForImage waits for the given Image ID to become ready.
func WaitForImage(client *gophercloud.ServiceClient, imageId string) error {
	maxNumErrors := 10
	numErrors := 0

	for {
		image, err := images.Get(client, imageId).Extract()
		if err != nil {
			errCode, ok := err.(*gophercloud.UnexpectedResponseCodeError)
			if ok && (errCode.Actual == 500 || errCode.Actual == 404) {
				numErrors++
				if numErrors >= maxNumErrors {
					log.Printf("[ERROR] Maximum number of errors (%d) reached; failing with: %s", numErrors, err)
					return err
				}
				log.Printf("[ERROR] %d error received, will ignore and retry: %s", errCode.Actual, err)
				time.Sleep(2 * time.Second)
				continue
			}

			return err
		}

		if image.Status == "ACTIVE" {
			return nil
		}

		log.Printf("Waiting for image creation status: %s (%d%%)", image.Status, image.Progress)
		time.Sleep(2 * time.Second)
	}
}

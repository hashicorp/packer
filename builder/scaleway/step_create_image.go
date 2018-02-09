package scaleway

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/scaleway/scaleway-cli/pkg/api"
)

type stepImage struct{}

func (s *stepImage) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*api.ScalewayAPI)
	ui := state.Get("ui").(packer.Ui)
	c := state.Get("config").(Config)
	snapshotID := state.Get("snapshot_id").(string)
	bootscriptID := ""

	ui.Say(fmt.Sprintf("Creating image: %v", c.ImageName))

	image, err := client.GetImage(c.Image)
	if err != nil {
		err := fmt.Errorf("Error getting initial image info: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if image.DefaultBootscript != nil {
		bootscriptID = image.DefaultBootscript.Identifier
	}

	imageID, err := client.PostImage(snapshotID, c.ImageName, bootscriptID, image.Arch)
	if err != nil {
		err := fmt.Errorf("Error creating image: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	log.Printf("Image ID: %s", imageID)
	state.Put("image_id", imageID)
	state.Put("image_name", c.ImageName)
	state.Put("region", c.Region)

	return multistep.ActionContinue
}

func (s *stepImage) Cleanup(state multistep.StateBag) {
	// no cleanup
}

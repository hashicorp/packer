package scaleway

import (
	"context"
	"fmt"
	"log"

	"github.com/dustin/go-humanize"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/scaleway/scaleway-cli/pkg/api"
)

type stepImage struct{}

func (s *stepImage) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*api.ScalewayAPI)
	ui := state.Get("ui").(packer.Ui)
	c := state.Get("config").(*Config)
	snapshotID := state.Get("snapshot_id").(string)
	bootscriptID := ""
	arch := ""

	ui.Say(fmt.Sprintf("Creating image: %v", c.ImageName))

	_, err := humanize.ParseBytes(c.Image)
	if err != nil {
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
		arch = image.Arch
	} else {
		// default to bootscript arch to find the arch of the image
		if c.Bootscript != "" {
			bootscripts, err := client.ResolveBootscript(c.Bootscript)
			if err != nil || len(bootscripts) == 0 {
				err := fmt.Errorf("Error getting arch from bootscript: %s", err)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}
			// pick the first one anyways
			arch = bootscripts[0].Arch
		}
	}

	imageID, err := client.PostImage(snapshotID, c.ImageName, bootscriptID, arch)
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

package classic

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-oracle-terraform/compute"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepListImages struct{}

func (s *stepListImages) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	// get variables from state
	ui := state.Get("ui").(packer.Ui)
	config := state.Get("config").(*Config)
	client := state.Get("client").(*compute.ComputeClient)
	ui.Say("Adding image to image list...")

	// TODO: Try to get image list
	imageListClient := client.ImageList()
	getInput := compute.GetImageListInput{
		Name: config.DestImageList,
	}
	imList, err := imageListClient.GetImageList(&getInput)
	if err != nil {
		ui.Say(fmt.Sprintf(err.Error()))
		// If the list didn't exist, create it.
		ui.Say(fmt.Sprintf("Creating image list: %s", config.DestImageList))

		ilInput := compute.CreateImageListInput{
			Name:        config.DestImageList,
			Description: "Packer-built image list",
		}

		// Load the packer-generated SSH key into the Oracle Compute cloud.
		imList, err = imageListClient.CreateImageList(&ilInput)
		if err != nil {
			err = fmt.Errorf("Problem creating an image list through Oracle's API: %s", err)
			ui.Error(err.Error())
			state.Put("error", err)
			return multistep.ActionHalt
		}
		ui.Message(fmt.Sprintf("Image list %s created!", imList.URI))
	}

	// Now create and image list entry for the image into that list.
	snap := state.Get("snapshot").(*compute.Snapshot)
	entriesClient := client.ImageListEntries()
	entriesInput := compute.CreateImageListEntryInput{
		Name:          config.DestImageList,
		MachineImages: []string{fmt.Sprintf("Compute-%s/%s/%s", config.IdentityDomain, config.Username, snap.MachineImage)},
		Version:       1,
	}
	entryInfo, err := entriesClient.CreateImageListEntry(&entriesInput)
	if err != nil {
		err = fmt.Errorf("Problem creating an image list entry: %s", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}
	ui.Message(fmt.Sprintf("created image list entry %s", entryInfo.Name))
	return multistep.ActionContinue
}

func (s *stepListImages) Cleanup(state multistep.StateBag) {
	// Nothing to do
	return
}

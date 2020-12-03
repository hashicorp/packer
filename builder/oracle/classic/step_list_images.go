package classic

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-oracle-terraform/compute"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type stepListImages struct{}

func (s *stepListImages) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	// get variables from state
	ui := state.Get("ui").(packersdk.Ui)
	config := state.Get("config").(*Config)
	client := state.Get("client").(*compute.Client)
	ui.Say("Adding image to image list...")

	imageListClient := client.ImageList()
	getInput := compute.GetImageListInput{
		Name: config.DestImageList,
	}
	imList, err := imageListClient.GetImageList(&getInput)
	if err != nil {
		// If the list didn't exist, create it.
		ui.Say(fmt.Sprintf(err.Error()))
		ui.Say(fmt.Sprintf("Destination image list %s does not exist; Creating it...",
			config.DestImageList))

		ilInput := compute.CreateImageListInput{
			Name:        config.DestImageList,
			Description: "Packer-built image list",
		}

		imList, err = imageListClient.CreateImageList(&ilInput)
		if err != nil {
			err = fmt.Errorf("Problem creating image list: %s", err)
			ui.Error(err.Error())
			state.Put("error", err)
			return multistep.ActionHalt
		}
		ui.Message(fmt.Sprintf("Image list %s created!", imList.URI))
	}

	// Now create and image list entry for the image into that list.
	machineImage := state.Get("machine_image").(string)
	version := len(imList.Entries) + 1
	entriesClient := client.ImageListEntries()
	entriesInput := compute.CreateImageListEntryInput{
		Name:          config.DestImageList,
		MachineImages: []string{config.Identifier(machineImage)},
		Version:       version,
	}
	entryInfo, err := entriesClient.CreateImageListEntry(&entriesInput)
	if err != nil {
		err = fmt.Errorf("Problem creating an image list entry: %s", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}
	state.Put("image_list_entry", entryInfo)
	ui.Message(fmt.Sprintf("created image list entry %s", entryInfo.Name))

	machineImagesClient := client.MachineImages()
	getImagesInput := compute.GetMachineImageInput{
		Name: config.ImageName,
	}

	// Update image list default to use latest version
	updateInput := compute.UpdateImageListInput{
		Default:     version,
		Description: config.DestImageListDescription,
		Name:        config.DestImageList,
	}
	_, err = imageListClient.UpdateImageList(&updateInput)
	if err != nil {
		err = fmt.Errorf("Problem updating default image list version: %s", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

	// Grab info about the machine image to return with the artifact
	imInfo, err := machineImagesClient.GetMachineImage(&getImagesInput)
	if err != nil {
		err = fmt.Errorf("Problem getting machine image info: %s", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}
	state.Put("machine_image_file", imInfo.File)
	state.Put("machine_image_name", imInfo.Name)
	state.Put("image_list_version", version)

	return multistep.ActionContinue
}

func (s *stepListImages) Cleanup(state multistep.StateBag) {
	// Nothing to do
	return
}

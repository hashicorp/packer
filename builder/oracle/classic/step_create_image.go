package classic

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/go-oracle-terraform/compute"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type stepCreateImage struct {
}

func (s *stepCreateImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	client := state.Get("client").(*compute.Client)
	config := state.Get("config").(*Config)
	imageFile := state.Get("image_file").(string)

	// Image uploaded, let's register it
	machineImageClient := client.MachineImages()
	createMI := &compute.CreateMachineImageInput{
		// Two-part name of the account
		Account:     fmt.Sprintf("/Compute-%s/cloud_storage", config.IdentityDomain),
		Description: "Packer generated Machine Image.",
		// The three-part name of the object
		Name: config.ImageName,
		// image_file.tar.gz, where image_file is the .tar.gz name of the
		// machine image file that you have uploaded to Oracle Cloud
		// Infrastructure Object Storage Classic.
		File: imageFile,
	}
	mi, err := machineImageClient.CreateMachineImage(createMI)
	if err != nil {
		err = fmt.Errorf("Error creating machine image: %s", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}
	log.Println("Registered machine image.")
	state.Put("machine_image", mi.Name)

	return multistep.ActionContinue
}

func (s *stepCreateImage) Cleanup(state multistep.StateBag) {
	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)
	if !cancelled && !halted {
		return
	}

	client := state.Get("client").(*compute.Client)
	config := state.Get("config").(*Config)

	ui := state.Get("ui").(packersdk.Ui)
	ui.Say("Cleaning up Image...")

	machineImageClient := client.MachineImages()
	deleteMI := &compute.DeleteMachineImageInput{
		Name: config.ImageName,
	}

	if err := machineImageClient.DeleteMachineImage(deleteMI); err != nil {
		ui.Error(fmt.Sprintf("Error cleaning up machine image: %s", err))
		return
	}

	ui.Message(fmt.Sprintf("Deleted Image: %s", config.ImageName))
}

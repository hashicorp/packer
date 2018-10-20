package classic

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/go-oracle-terraform/compute"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepCreateImage struct {
}

func (s *stepCreateImage) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	client := state.Get("client").(*compute.ComputeClient)
	config := state.Get("config").(*Config)
	imageFile := state.Get("image_file").(string)

	// Image uploaded, let's register it
	machineImageClient := client.MachineImages()
	createMI := &compute.CreateMachineImageInput{
		// Two-part name of the account
		Account:     fmt.Sprintf("/Compute-%s/cloud_storage", config.IdentityDomain),
		Description: "Packer generated TODO",
		// The three-part name of the object
		Name: config.ImageName,
		// image_file.tar.gz, where image_file is the .tar.gz name of the machine image file that you have uploaded to Oracle Cloud Infrastructure Object Storage Classic.
		File: imageFile,
	}
	log.Printf("CreateMachineImageInput: %+v", createMI)
	mi, err := machineImageClient.CreateMachineImage(createMI)
	if err != nil {
		err = fmt.Errorf("Error creating machine image: %s, %+v", err, mi)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}
	log.Printf("Registered machine image: %+v", mi)
	state.Put("machine_image", mi.Name)
	/*
		Registered machine image: &{
			Account:/Compute-ptstest/cloud_storage
			Attributes:map[]
			Audited:
			Description:Packer generated TODO
			ErrorReason: Hypervisor:map[mode:hvm]
			ImageFormat:raw
			File:mwhooker_test_1539898463.tar.gz
			Name:mwhooker_test_1539898463
			NoUpload:true
			Platform:linux
			Sizes:map[uploaded:5.79793509e+08 total:5.79793509e+08 decompressed:1.610612736e+10]
			State:available
			URI:https://api-z61.compute.us6.oraclecloud.com/machineimage/Compute-ptstest/mhooker%40hashicorp.com/mwhooker_test_1539898463
		}
	*/
	/* TODO:
	*	POST /machineimage/ DONE
		POST /imagelist/ DONE
		POST /imagelistentry/ DONE
		in that order.
	* re-use step_list_images DONE
	* Documentation
	* Configuration (master/builder images & entry, destination stuff, etc)
		* Image entry for both master/builder
		https://github.com/hashicorp/packer/issues/6833
	* split master/builder image/connection config. i.e. build anything, master only linux
	* correct artifact DONE
	* Cleanup this step
	*/

	return multistep.ActionContinue
}

func (s *stepCreateImage) Cleanup(state multistep.StateBag) {
	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)
	if !cancelled && !halted {
		return
	}

	client := state.Get("client").(*compute.ComputeClient)
	config := state.Get("config").(*Config)

	ui := state.Get("ui").(packer.Ui)
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

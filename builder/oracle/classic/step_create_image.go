package classic

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/go-oracle-terraform/compute"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

type stepCreateImage struct {
	uploadImageCommand string
	tempContainer      string
	imageName          string
}

type uploadCmdData struct {
	Username  string
	Password  string
	AccountID string
	Container string
	ImageName string
}

func (s *stepCreateImage) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	//hook := state.Get("hook").(packer.Hook)
	ui := state.Get("ui").(packer.Ui)
	comm := state.Get("communicator").(packer.Communicator)
	client := state.Get("client").(*compute.ComputeClient)
	config := state.Get("config").(*Config)

	config.ctx.Data = uploadCmdData{
		Username:  config.Username,
		Password:  config.Password,
		AccountID: config.IdentityDomain,
		Container: s.tempContainer,
		ImageName: s.imageName,
	}
	uploadImageCmd, err := interpolate.Render(s.uploadImageCommand, &config.ctx)
	if err != nil {
		err := fmt.Errorf("Error processing image upload command: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	command := fmt.Sprintf(`#!/bin/sh
	set -e
	set -x
	mkdir /builder
	mkfs -t ext3 /dev/xvdb
	mount /dev/xvdb /builder
	chown opc:opc /builder
	cd /builder
	dd if=/dev/xvdc bs=8M status=progress | cp --sparse=always /dev/stdin diskimage.raw
	tar czSf ./diskimage.tar.gz ./diskimage.raw
	rm diskimage.raw
	%s`, uploadImageCmd)

	dest := "/tmp/create-packer-diskimage.sh"
	comm.Upload(dest, strings.NewReader(command), nil)
	cmd := &packer.RemoteCmd{
		Command: fmt.Sprintf("sudo /bin/sh %s", dest),
	}
	if err := cmd.StartWithUi(comm, ui); err != nil {
		err = fmt.Errorf("Problem creating image`: %s", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

	if cmd.ExitStatus != 0 {
		err = fmt.Errorf("Create Disk Image command failed with exit code %d", cmd.ExitStatus)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

	// Image uploaded, let's register it
	machineImageClient := client.MachineImages()
	createMI := &compute.CreateMachineImageInput{
		// Two-part name of the account
		Account:     fmt.Sprintf("/Compute-%s/cloud_storage", config.IdentityDomain),
		Description: "Packer generated TODO",
		// The three-part name of the object
		Name: config.Identifier(s.imageName),
		// image_file.tar.gz, where image_file is the .tar.gz name of the machine image file that you have uploaded to Oracle Cloud Infrastructure Object Storage Classic.
		File: fmt.Sprintf("%s.tar.gz", s.imageName),
	}
	log.Printf("CreateMachineImageInput: %+v", createMI)
	mi, err := machineImageClient.CreateMachineImage(createMI)
	if err != nil {
		err = fmt.Errorf("Error creating machine image: %s", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}
	log.Printf("Registered machine image: %+v", mi)
	/* TODO:
	1. POST /machineimage/, POST /imagelist/, and POST /imagelistentry/ methods, in that order.
	2. re-use step_list_images
	3. Don't push commits with passwords in them
	4. Documentation
	5. Configuration (master/builder images & entry, destination stuff, etc)
	6. split master/builder image/connection config. i.e. build anything, master only linux
	7. correct artifact
	8. hide password from logs
	*/
	//machineImageClient.CreateMachineImage()

	return multistep.ActionContinue
}

func (s *stepCreateImage) Cleanup(state multistep.StateBag) {}

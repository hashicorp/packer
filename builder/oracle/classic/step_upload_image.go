package classic

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/interpolate"
)

type stepUploadImage struct {
	UploadImageCommand string
	segmentPath        string
}

type uploadCmdData struct {
	Username    string
	Password    string
	AccountID   string
	ImageFile   string
	SegmentPath string
}

func (s *stepUploadImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	comm := state.Get("communicator").(packer.Communicator)
	config := state.Get("config").(*Config)
	runID := state.Get("run_id").(string)

	imageFile := fmt.Sprintf("%s.tar.gz", config.ImageName)
	state.Put("image_file", imageFile)
	s.segmentPath = fmt.Sprintf("compute_images_segments/%s/_segment_/%s", imageFile, runID)

	config.ctx.Data = uploadCmdData{
		Username:    config.Username,
		Password:    config.Password,
		AccountID:   config.IdentityDomain,
		ImageFile:   imageFile,
		SegmentPath: s.segmentPath,
	}
	uploadImageCmd, err := interpolate.Render(s.UploadImageCommand, &config.ctx)
	if err != nil {
		err := fmt.Errorf("Error processing image upload command: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	command := fmt.Sprintf(`#!/bin/sh
	set -e
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
	if err := cmd.RunWithUi(ctx, comm, ui); err != nil {
		err = fmt.Errorf("Problem creating image`: %s", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

	if cmd.ExitStatus() != 0 {
		err = fmt.Errorf("Create Disk Image command failed with exit code %d", cmd.ExitStatus())
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

	ui.Say("Uploaded image to object storage.")

	return multistep.ActionContinue
}

func (s *stepUploadImage) Cleanup(state multistep.StateBag) {
	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)
	if !cancelled && !halted {
		return
	}
	log.Printf("Some segments may need to be manually cleaned at '%s'", s.segmentPath)
}

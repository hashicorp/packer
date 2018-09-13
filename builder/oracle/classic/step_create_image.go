package classic

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

type stepCreateImage struct {
	uploadImageCommand string
}

type uploadCmdData struct {
	DiskImagePath string
}

func (s *stepCreateImage) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	//hook := state.Get("hook").(packer.Hook)
	ui := state.Get("ui").(packer.Ui)
	comm := state.Get("communicator").(packer.Communicator)
	config := state.Get("config").(*Config)

	config.ctx.Data = uploadCmdData{
		DiskImagePath: "./diskimage.tar.gz",
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
	mkdir /builder
	mkfs -t ext3 /dev/xvdb
	mount /dev/xvdb /builder
	chown opc:opc /builder
	cd /builder
	dd if=/dev/xvdc bs=8M status=progress | cp --sparse=always /dev/stdin diskimage.raw
	tar czSf ./diskimage.tar.gz ./diskimage.raw
	%s`, uploadImageCmd)

	dest := "/tmp/create-packer-diskimage.sh"
	comm.Upload(dest, strings.NewReader(command), nil)
	cmd := &packer.RemoteCmd{
		Command: fmt.Sprintf("sudo /bin/sh %s", dest),
	}
	if err := cmd.StartWithUi(comm, ui); err != nil {
		err = fmt.Errorf("Problem creating volume: %s", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepCreateImage) Cleanup(state multistep.StateBag) {}

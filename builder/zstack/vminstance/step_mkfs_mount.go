package vminstance

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/packer/builder/zstack/zstacktype"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

type StepMkfsMount struct {
}

func (s *StepMkfsMount) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	_, config, ui := GetCommonFromState(state)
	comm := state.Get("communicator").(packer.Communicator)
	ui.Say("start mkfs and mount...")

	volume := state.Get(DataVolume).(*zstacktype.DataVolume)
	vm := state.Get(Vm).(*zstacktype.VmInstance)
	if volume == nil || vm == nil {
		return multistep.ActionContinue
	} else {
		ui.Say(fmt.Sprintf("volume: %s", volume.Uuid))
		ui.Say(fmt.Sprintf("vm: %s", vm.Uuid))
	}

	var command string
	if config.DataVolumeSize != "" {
		command = fmt.Sprintf(`#!/bin/sh
		mkdir %s
		mkfs -t %s /dev/vdb
		mount /dev/vdb %s
		`, config.MountPath, config.FileSystemType, config.MountPath)
	} else {
		command = fmt.Sprintf(`#!/bin/sh
		mkdir -p %s
		mount /dev/vdb %s
		`, config.MountPath, config.MountPath)
	}

	dest, err := interpolate.Render("/tmp/mkfs-mount-{{timestamp}}.sh", nil)
	if err != nil {
		return halt(state, err, "")
	}
	comm.Upload(dest, strings.NewReader(command), nil)
	cmd := &packer.RemoteCmd{
		Command: fmt.Sprintf("/bin/sh %s", dest),
	}
	if err := cmd.RunWithUi(ctx, comm, ui); err != nil {
		err = fmt.Errorf("Problem fdisk and mount`: %s", err.Error())
		return halt(state, err, "")
	}
	if cmd.ExitStatus() != 0 {
		return halt(state, fmt.Errorf("fdisk or mount failed with exit code %d", cmd.ExitStatus()), "")
	}

	return multistep.ActionContinue
}

func (s *StepMkfsMount) Cleanup(state multistep.StateBag) {
	_, _, ui := GetCommonFromState(state)
	ui.Say("cleanup mkfs and mount executing...")
}

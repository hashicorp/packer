package classic

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepCreateImage struct{}

func (s *stepCreateImage) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	//hook := state.Get("hook").(packer.Hook)
	ui := state.Get("ui").(packer.Ui)
	comm := state.Get("communicator").(packer.Communicator)
	command := `#!/bin/sh
	set -e
	mkdir /builder
	mkfs -t ext3 /dev/xvdb
	mount /dev/xvdb /builder
	chown opc:opc /builder
	cd /builder
	dd if=/dev/xvdc bs=8M status=progress | cp --sparse=always /dev/stdin diskimage.raw
	tar czSf ./diskimage.tar.gz ./diskimage.raw`

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

package classic

import (
	"context"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepCreateImage struct{}

func (s *stepCreateImage) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	//hook := state.Get("hook").(packer.Hook)
	ui := state.Get("ui").(packer.Ui)
	comm := state.Get("communicator").(packer.Communicator)
	commands := []string{
		"mkdir ./builder",
		"sudo mkfs -t ext3 /dev/xvdb",
		"sudo mount /dev/xvdb ./builder",
		"sudo chown opc:opc ./builder",
		"cd ./builder",
		"sudo dd if=/dev/xvdc bs=8M status=progress | cp --sparse=always /dev/stdin diskimage.raw",
		"tar czSf ./diskimage.tar.gz ./diskimage.raw",
	}
	for _, c := range commands {
		cmd := packer.RemoteCmd{
			Command: c,
		}
		cmd.StartWithUi(comm, ui)
	}
	//	comm.Start("

	/*
		// Provision
		log.Println("Running the provision hook")
		if err := hook.Run(packer.HookProvision, ui, comm, nil); err != nil {
			state.Put("error", err)
			return multistep.ActionHalt
		}
	*/

	return multistep.ActionContinue
}

func (s *stepCreateImage) Cleanup(state multistep.StateBag) {}

package hyperone

import (
	"context"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep/commonsteps"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type stepChrootProvision struct{}

func (s *stepChrootProvision) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	wrappedCommand := state.Get("wrappedCommand").(CommandWrapper)
	sshCommunicator := state.Get("communicator").(packersdk.Communicator)

	comm := &ChrootCommunicator{
		Chroot:     config.ChrootMountPath,
		CmdWrapper: wrappedCommand,
		Wrapped:    sshCommunicator,
	}

	stepProvision := commonsteps.StepProvision{Comm: comm}
	return stepProvision.Run(ctx, state)
}

func (s *stepChrootProvision) Cleanup(multistep.StateBag) {}

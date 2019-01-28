package hyperone

import (
	"context"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepChrootProvision struct{}

func (s *stepChrootProvision) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	wrappedCommand := state.Get("wrappedCommand").(CommandWrapper)
	sshCommunicator := state.Get("communicator").(packer.Communicator)

	comm := &ChrootCommunicator{
		Chroot:     config.ChrootMountPath,
		CmdWrapper: wrappedCommand,
		Wrapped:    sshCommunicator,
	}

	stepProvision := common.StepProvision{Comm: comm}
	return stepProvision.Run(ctx, state)
}

func (s *stepChrootProvision) Cleanup(multistep.StateBag) {}

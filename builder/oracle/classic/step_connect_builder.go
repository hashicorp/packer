package classic

import (
	"context"

	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
)

type stepConnectBuilder struct {
	*communicator.StepConnectSSH
	KeyName string
}

func (s *stepConnectBuilder) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	s.Config.SSHKeyPairName = s.KeyName
	return s.StepConnectSSH.Run(ctx, state)
}

func (s *stepConnectBuilder) Cleanup(state multistep.StateBag) {
	s.StepConnectSSH.Cleanup(state)
}

package dtl

import (
	"context"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type StepSaveWinRMPassword struct {
	Password  string
	BuildName string
}

func (s *StepSaveWinRMPassword) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	// store so that we can access this later during provisioning
	state.Put("winrm_password", s.Password)
	packer.LogSecretFilter.Set(s.Password)
	return multistep.ActionContinue
}

func (s *StepSaveWinRMPassword) Cleanup(multistep.StateBag) {}

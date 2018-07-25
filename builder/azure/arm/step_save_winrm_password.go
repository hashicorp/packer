package arm

import (
	"context"

	commonhelper "github.com/hashicorp/packer/helper/common"
	"github.com/hashicorp/packer/helper/multistep"
)

type StepSaveWinRMPassword struct {
	Password  string
	BuildName string
}

func (s *StepSaveWinRMPassword) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	// store so that we can access this later during provisioning
	commonhelper.SetSharedState("winrm_password", s.Password, s.BuildName)
	return multistep.ActionContinue
}

func (s *StepSaveWinRMPassword) Cleanup(multistep.StateBag) {
	commonhelper.RemoveSharedStateFile("winrm_password", s.BuildName)
}

package arm

import (
	"context"

	commonhelper "github.com/hashicorp/packer/helper/common"
	"github.com/hashicorp/packer/helper/multistep"
)

type StepSaveWinRMPassword struct {
	Password string
}

func (s *StepSaveWinRMPassword) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	// store so that we can access this later during provisioning
	commonhelper.SetSharedState("winrm_password", s.Password)
	return multistep.ActionContinue
}

func (s *StepSaveWinRMPassword) Cleanup(multistep.StateBag) {
	commonhelper.RemoveSharedStateFile("winrm_password")
}

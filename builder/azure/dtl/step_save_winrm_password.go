package dtl

import (
	"context"

	"github.com/hashicorp/packer/builder/azure/common/constants"
	commonhelper "github.com/hashicorp/packer/helper/common"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type StepSaveWinRMPassword struct {
	Password  string
	BuildName string
}

func (s *StepSaveWinRMPassword) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	// store so that we can access this later during provisioning
	err := commonhelper.SetSharedState("winrm_password", s.Password, s.BuildName)
	if err != nil {
		state.Put(constants.Error, err)

		return multistep.ActionHalt
	}
	packer.LogSecretFilter.Set(s.Password)
	return multistep.ActionContinue
}

func (s *StepSaveWinRMPassword) Cleanup(multistep.StateBag) {
	commonhelper.RemoveSharedStateFile("winrm_password", s.BuildName)
}

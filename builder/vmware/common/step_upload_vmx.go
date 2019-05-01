package common

import (
	"context"

	"github.com/hashicorp/packer/helper/multistep"
)

// This step upload the VMX to the remote host
//
// Produces:
//   <nothing>
type StepUploadVMX struct {
}

func (c *StepUploadVMX) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	return multistep.ActionContinue
}

func (StepUploadVMX) Cleanup(multistep.StateBag) {}

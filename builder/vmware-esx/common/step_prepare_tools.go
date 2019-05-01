package common

import (
	"context"

	"github.com/hashicorp/packer/helper/multistep"
)

type StepPrepareTools struct {
	ToolsUploadFlavor string
}

func (c *StepPrepareTools) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {

	// XXX: The original ESX driver skips this step via multistep.ActionContinue
	return multistep.ActionContinue
}

func (c *StepPrepareTools) Cleanup(multistep.StateBag) {}

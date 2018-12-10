package common

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/packer/helper/multistep"
)

type StepPrepareTools struct {
	ToolsUploadFlavor string
}

func (c *StepPrepareTools) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)

	// XXX: The original ESX driver skips this step via multistep.ActionContinue
	return multistep.ActionContinue
}

func (c *StepPrepareTools) Cleanup(multistep.StateBag) {}

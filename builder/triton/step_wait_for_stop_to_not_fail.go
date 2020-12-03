package triton

import (
	"context"
	"time"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

// StepWaitForStopNotToFail waits for 10 seconds before returning with continue
// in order to prevent an observed issue where machines stopped immediately after
// they are started never actually stop.
type StepWaitForStopNotToFail struct{}

func (s *StepWaitForStopNotToFail) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	ui.Say("Waiting 10 seconds to avoid potential SDC bug...")
	time.Sleep(10 * time.Second)
	return multistep.ActionContinue
}

func (s *StepWaitForStopNotToFail) Cleanup(state multistep.StateBag) {
	// No clean up required...
}

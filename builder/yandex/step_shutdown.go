package yandex

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
)

type stepShutdown struct {
	Debug bool
}

func (s *stepShutdown) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	sdk := state.Get("sdk").(*ycsdk.SDK)
	ui := state.Get("ui").(packer.Ui)
	instanceID := state.Get("instance_id").(string)

	// Gracefully power off the instance. We have to retry this a number
	// of times because sometimes it says it completed when it actually
	// did absolutely nothing (*ALAKAZAM!* magic!). We give up after
	// a pretty arbitrary amount of time.
	ui.Say("Gracefully shutting down instance...")
	op, err := sdk.WrapOperation(sdk.Compute().Instance().Stop(context.Background(), &compute.StopInstanceRequest{
		InstanceId: instanceID,
	}))
	if err != nil {
		return stepHaltWithError(state, fmt.Errorf("Error shutting down instance: %s", err))
	}
	err = op.Wait(context.Background())
	if err != nil {
		return stepHaltWithError(state, fmt.Errorf("Error shutting down instance: %s", err))
	}

	if s.Debug {
		ui.Message("Instance status before image create:")
		displayInstanceStatus(sdk, instanceID, ui)
	}

	return multistep.ActionContinue
}

func (s *stepShutdown) Cleanup(state multistep.StateBag) {
	// no cleanup
}

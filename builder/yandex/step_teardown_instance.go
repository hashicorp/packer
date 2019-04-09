package yandex

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
)

type stepTeardownInstance struct{}

func (s *stepTeardownInstance) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	sdk := state.Get("sdk").(*ycsdk.SDK)
	ui := state.Get("ui").(packer.Ui)
	c := state.Get("config").(*Config)

	instanceID := state.Get("instance_id").(string)

	ui.Say("Deleting instance...")
	ctx, cancel := context.WithTimeout(ctx, c.StateTimeout)
	defer cancel()

	op, err := sdk.WrapOperation(sdk.Compute().Instance().Delete(ctx, &compute.DeleteInstanceRequest{
		InstanceId: instanceID,
	}))
	if err != nil {
		return stepHaltWithError(state, fmt.Errorf("Error deleting instance: %s", err))
	}
	err = op.Wait(ctx)
	if err != nil {
		return stepHaltWithError(state, fmt.Errorf("Error deleting instance: %s", err))
	}

	ui.Message("Instance has been deleted!")
	state.Put("instance_id", "")

	return multistep.ActionContinue
}

func (s *stepTeardownInstance) Cleanup(state multistep.StateBag) {
	// no cleanup
}

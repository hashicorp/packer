package yandex

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
)

type StepTeardownInstance struct {
	SerialLogFile string
}

func (s *StepTeardownInstance) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	sdk := state.Get("sdk").(*ycsdk.SDK)
	ui := state.Get("ui").(packersdk.Ui)
	c := state.Get("config").(*Config)

	instanceID := state.Get("instance_id").(string)

	ui.Say("Stopping instance...")
	ctx, cancel := context.WithTimeout(ctx, c.StateTimeout)
	defer cancel()

	if s.SerialLogFile != "" {
		err := writeSerialLogFile(ctx, state, s.SerialLogFile)
		if err != nil {
			ui.Error(err.Error())
		}
	}

	op, err := sdk.WrapOperation(sdk.Compute().Instance().Stop(ctx, &compute.StopInstanceRequest{
		InstanceId: instanceID,
	}))
	if err != nil {
		return StepHaltWithError(state, fmt.Errorf("Error stopping instance: %s", err))
	}
	err = op.Wait(ctx)
	if err != nil {
		return StepHaltWithError(state, fmt.Errorf("Error stopping instance: %s", err))
	}

	ui.Say("Deleting instance...")
	op, err = sdk.WrapOperation(sdk.Compute().Instance().Delete(ctx, &compute.DeleteInstanceRequest{
		InstanceId: instanceID,
	}))
	if err != nil {
		return StepHaltWithError(state, fmt.Errorf("Error deleting instance: %s", err))
	}
	err = op.Wait(ctx)
	if err != nil {
		return StepHaltWithError(state, fmt.Errorf("Error deleting instance: %s", err))
	}

	ui.Message("Instance has been deleted!")
	state.Put("instance_id", "")

	return multistep.ActionContinue
}

func (s *StepTeardownInstance) Cleanup(state multistep.StateBag) {
	// no cleanup
}

package yandex

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/retry"
)

const CloudInitScriptStatusKey = "cloud-init-status"
const StartupScriptStatusError = "cloud-init-error"
const StartupScriptStatusDone = "cloud-init-done"

type StepWaitCloudInitScript int

// Run reads the instance metadata and looks for the log entry
// indicating the cloud-init script finished.
func (s *StepWaitCloudInitScript) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	_ = state.Get("config").(*Config)
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packersdk.Ui)
	instanceID := state.Get("instance_id").(string)

	ui.Say("Waiting for any running cloud-init script to finish...")

	// Keep checking the serial port output to see if the cloud-init script is done.
	err := retry.Config{
		ShouldRetry: func(error) bool {
			return true
		},
		RetryDelay: (&retry.Backoff{InitialBackoff: 10 * time.Second, MaxBackoff: 60 * time.Second, Multiplier: 2}).Linear,
	}.Run(ctx, func(ctx context.Context) error {
		status, err := driver.GetInstanceMetadata(ctx, instanceID, CloudInitScriptStatusKey)

		if err != nil {
			err := fmt.Errorf("Error getting cloud-init script status: %s", err)
			return err
		}

		if status == StartupScriptStatusError {
			err = errors.New("Cloud-init script error.")
			return err
		}

		done := status == StartupScriptStatusDone
		if !done {
			ui.Say("Cloud-init script not finished yet. Waiting...")
			return errors.New("Cloud-init script not done.")
		}

		return nil
	})

	if err != nil {
		err := fmt.Errorf("Error waiting for cloud-init script to finish: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	ui.Say("Cloud-init script has finished running.")
	return multistep.ActionContinue
}

// Cleanup.
func (s *StepWaitCloudInitScript) Cleanup(state multistep.StateBag) {}

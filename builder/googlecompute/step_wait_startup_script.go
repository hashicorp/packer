package googlecompute

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/retry"
)

type StepWaitStartupScript int

// Run reads the instance metadata and looks for the log entry
// indicating the startup script finished.
func (s *StepWaitStartupScript) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packersdk.Ui)
	instanceName := state.Get("instance_name").(string)

	if config.WrapStartupScriptFile.False() {
		return multistep.ActionContinue
	}

	ui.Say("Waiting for any running startup script to finish...")
	// Keep checking the serial port output to see if the startup script is done.
	err := retry.Config{
		ShouldRetry: func(error) bool {
			return true
		},
		RetryDelay: (&retry.Backoff{InitialBackoff: 10 * time.Second, MaxBackoff: 60 * time.Second, Multiplier: 2}).Linear,
	}.Run(ctx, func(ctx context.Context) error {
		status, err := driver.GetInstanceMetadata(config.Zone,
			instanceName, StartupScriptStatusKey)

		if err != nil {
			err := fmt.Errorf("Error getting startup script status: %s", err)
			return err
		}

		if status == StartupScriptStatusError {
			err = errors.New("Startup script error.")
			return err
		}

		done := status == StartupScriptStatusDone
		if !done {
			ui.Say("Startup script not finished yet. Waiting...")
			return errors.New("Startup script not done.")
		}

		return nil
	})

	if err != nil {
		err := fmt.Errorf("Error waiting for startup script to finish: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	ui.Say("Startup script, if any, has finished running.")
	return multistep.ActionContinue
}

// Cleanup.
func (s *StepWaitStartupScript) Cleanup(state multistep.StateBag) {}

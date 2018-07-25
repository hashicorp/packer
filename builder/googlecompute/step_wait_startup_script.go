package googlecompute

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type StepWaitStartupScript int

// Run reads the instance metadata and looks for the log entry
// indicating the startup script finished.
func (s *StepWaitStartupScript) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
	instanceName := state.Get("instance_name").(string)

	ui.Say("Waiting for any running startup script to finish...")

	// Keep checking the serial port output to see if the startup script is done.
	err := common.Retry(10, 60, 0, func(_ uint) (bool, error) {
		status, err := driver.GetInstanceMetadata(config.Zone,
			instanceName, StartupScriptStatusKey)

		if err != nil {
			err := fmt.Errorf("Error getting startup script status: %s", err)
			return false, err
		}

		if status == StartupScriptStatusError {
			err = errors.New("Startup script error.")
			return false, err
		}

		done := status == StartupScriptStatusDone
		if !done {
			ui.Say("Startup script not finished yet. Waiting...")
		}

		return done, nil
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

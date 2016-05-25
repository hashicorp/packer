package googlecompute

import(
	"fmt"
	"strings"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type StepWaitInstanceStartup int

// Run reads the instance serial port output and looks for the log entry indicating the startup script finished.
func (s *StepWaitInstanceStartup) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
	instanceName := state.Get("instance_name").(string)
	
	ui.Say("Waiting for any running startup script to finish...")

	// Keep checking the serial port output to see if the startup script is done.
	err := Retry(10, 60, 0, func() (bool, error) {
		output, err := driver.GetSerialPortOutput(config.Zone, instanceName)

		if err != nil {
			err := fmt.Errorf("Error getting serial port output: %s", err)
			return false, err
		}
		
		done := strings.Contains(output, StartupScriptDoneLog)
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
func (s *StepWaitInstanceStartup) Cleanup(state multistep.StateBag) {}

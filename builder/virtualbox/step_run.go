package virtualbox

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"time"
)

// This step starts the virtual machine.
//
// Uses:
//
// Produces:
type stepRun struct {
	vmName string
}

func (s *stepRun) Run(state map[string]interface{}) multistep.StepAction {
	config := state["config"].(*config)
	driver := state["driver"].(Driver)
	ui := state["ui"].(packer.Ui)
	vmName := state["vmName"].(string)

	ui.Say("Starting the virtual machine...")
	command := []string{"startvm", vmName, "--type", "gui"}
	if err := driver.VBoxManage(command...); err != nil {
		err := fmt.Errorf("Error starting VM: %s", err)
		state["error"] = err
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	s.vmName = vmName

	if int64(config.BootWait) > 0 {
		ui.Say(fmt.Sprintf("Waiting %s for boot...", config.BootWait))
		time.Sleep(config.BootWait)
	}

	return multistep.ActionContinue
}

func (s *stepRun) Cleanup(state map[string]interface{}) {
	if s.vmName == "" {
		return
	}

	driver := state["driver"].(Driver)
	ui := state["ui"].(packer.Ui)

	if running, _ := driver.IsRunning(s.vmName); running {
		if err := driver.VBoxManage("controlvm", s.vmName, "poweroff"); err != nil {
			ui.Error(fmt.Sprintf("Error shutting down VM: %s", err))
		}
	}
}

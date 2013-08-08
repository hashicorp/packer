package virtualbox

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"strings"
)

// This step executes additional VBoxManage commands as specified by the
// template.
//
// Uses:
//
// Produces:
type stepVBoxManage struct{}

func (s *stepVBoxManage) Run(state map[string]interface{}) multistep.StepAction {
	config := state["config"].(*config)
	driver := state["driver"].(Driver)
	ui := state["ui"].(packer.Ui)

	if len(config.VBoxManage) > 0 {
		ui.Say("Executing custom VBoxManage commands...")
	}

	for _, command := range config.VBoxManage {
		ui.Message(fmt.Sprintf("Executing: %s", strings.Join(command, " ")))
		if err := driver.VBoxManage(command...); err != nil {
			err := fmt.Errorf("Error executing command: %s", err)
			state["error"] = err
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (s *stepVBoxManage) Cleanup(state map[string]interface{}) {}

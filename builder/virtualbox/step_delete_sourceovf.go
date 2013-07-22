package virtualbox

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type stepDeleteSourceOvf struct{}

func (s *stepDeleteSourceOvf) Run(state map[string]interface{}) multistep.StepAction {
	config := state["config"].(*config)
	driver := state["driver"].(Driver)
	ui := state["ui"].(packer.Ui)

	name := config.VMName
	sourceVmName := fmt.Sprintf("source_%s", name)

	ui.Say("Unregistering and deleting source ovf ...")
	if err := driver.VBoxManage("unregistervm", sourceVmName, "--delete"); err != nil {
		ui.Error(fmt.Sprintf("Error deleting virtual machine: %s", err))
	}

	return multistep.ActionContinue
}

func (s *stepDeleteSourceOvf) Cleanup(state map[string]interface{}) {}

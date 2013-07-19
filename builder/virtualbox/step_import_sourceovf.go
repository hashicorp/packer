package virtualbox

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

// This step imports a source ovf image
//
// Produces:
//   sourceVmName string - The name of the VM imported from the ovf
type stepImportSourceOvf struct {
	sourceVmName string
}

func (s *stepImportSourceOvf) Run(state map[string]interface{}) multistep.StepAction {
	config := state["config"].(*config)
	driver := state["driver"].(Driver)
	ui := state["ui"].(packer.Ui)

	name := config.VMName
        sourceOvf := config.SourceOvfFile
	sourceVmName := fmt.Sprintf("source_%s", name)

	commands := make([][]string, 1)
	// determine the output path for the disk .. it defaults to your destdir at the moment.
	commands[0] = []string{
		"import", sourceOvf, "--vsys", "0",
		"--vmname", sourceVmName,
	}

	ui.Say("Importing source ovf...")
	for _, command := range commands {
		err := driver.VBoxManage(command...)
		if err != nil {
			err := fmt.Errorf("Error Importing Ovf: %s", err)
			state["error"] = err
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	// Set the source sourceVmName in the state bag so others can use it
	state["sourceVmName"] = s.sourceVmName

	return multistep.ActionContinue
}

func (s *stepImportSourceOvf) Cleanup(state map[string]interface{}) {
	if s.sourceVmName == "" {
		return
	}

	driver := state["driver"].(Driver)
	ui := state["ui"].(packer.Ui)

	ui.Say("Unregistering and deleting source ovf ...")
	if err := driver.VBoxManage("unregistervm", s.sourceVmName, "--delete"); err != nil {
		ui.Error(fmt.Sprintf("Error deleting virtual machine: %s", err))
	}
}

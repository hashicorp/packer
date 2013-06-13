package virtualbox

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"path/filepath"
)

// This step creates the virtual disk that will be used as the
// hard drive for the virtual machine.
//
// Uses:
//
// Produces:
//   exportPath string - The path to the resulting export.
type stepExport struct{}

func (s *stepExport) Run(state map[string]interface{}) multistep.StepAction {
	config := state["config"].(*config)
	driver := state["driver"].(Driver)
	ui := state["ui"].(packer.Ui)
	vmName := state["vmName"].(string)

	outputPath := filepath.Join(config.OutputDir, "packer.ovf")

	command := []string{
		"export",
		vmName,
		"--output",
		outputPath,
	}

	ui.Say("Exporting virtual machine...")
	err := driver.VBoxManage(command...)
	if err != nil {
		ui.Error(fmt.Sprintf("Error exporting virtual machine: %s", err))
		return multistep.ActionHalt
	}

	state["exportPath"] = outputPath

	return multistep.ActionContinue
}

func (s *stepExport) Cleanup(state map[string]interface{}) {}

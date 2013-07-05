package virtualbox

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"path/filepath"
)

// This step cleans up forwarded ports and exports the VM to an OVF.
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

	// Clear out the Packer-created forwarding rule
	ui.Say(fmt.Sprintf("Deleting forwarded port mapping for SSH (host port %d)", state["sshHostPort"]))
	command := []string{"modifyvm", vmName, "--natpf1", "delete", "packerssh"}
	if err := driver.VBoxManage(command...); err != nil {
		err := fmt.Errorf("Error deleting port forwarding rule: %s", err)
		state["error"] = err
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Export the VM to an OVF
	outputPath := filepath.Join(config.OutputDir, "packer.ovf")

	command = []string{
		"export",
		vmName,
		"--output",
		outputPath,
	}

	ui.Say("Exporting virtual machine...")
	err := driver.VBoxManage(command...)
	if err != nil {
		err := fmt.Errorf("Error exporting virtual machine: %s", err)
		state["error"] = err
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state["exportPath"] = outputPath

	return multistep.ActionContinue
}

func (s *stepExport) Cleanup(state map[string]interface{}) {}

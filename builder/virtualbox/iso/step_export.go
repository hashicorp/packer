package iso

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
	"path/filepath"
	"time"
)

// This step cleans up forwarded ports and exports the VM to an OVF.
//
// Uses:
//
// Produces:
//   exportPath string - The path to the resulting export.
type stepExport struct{}

func (s *stepExport) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*config)
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
	vmName := state.Get("vmName").(string)

	// Wait a second to ensure VM is really shutdown.
	log.Println("1 second timeout to ensure VM is really shutdown")
	time.Sleep(1 * time.Second)

	// Clear out the Packer-created forwarding rule
	ui.Say("Preparing to export machine...")
	ui.Message(fmt.Sprintf(
		"Deleting forwarded port mapping for SSH (host port %d)",
		state.Get("sshHostPort")))
	command := []string{"modifyvm", vmName, "--natpf1", "delete", "packerssh"}
	if err := driver.VBoxManage(command...); err != nil {
		err := fmt.Errorf("Error deleting port forwarding rule: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Export the VM to an OVF
	outputPath := filepath.Join(config.OutputDir, vmName+"."+config.Format)

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
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("exportPath", outputPath)

	return multistep.ActionContinue
}

func (s *stepExport) Cleanup(state multistep.StateBag) {}

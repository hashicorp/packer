package virtualbox

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"path/filepath"
	"strings"
	"time"
)

// This step creates the virtual disk that will be used as the
// hard drive for the virtual machine.
type stepCreateDisk struct{
	diskPath string
}

func (s *stepCreateDisk) Run(state map[string]interface{}) multistep.StepAction {
	config := state["config"].(*config)
	driver := state["driver"].(Driver)
	ui := state["ui"].(packer.Ui)

	format := "VDI"
	path := filepath.Join(config.OutputDir, fmt.Sprintf("%s.%s", config.VMName, strings.ToLower(format)))

	command := []string{
		"createhd",
		"--filename", path,
		"--size", "40",
		"--format", format,
		"--variant", "Standard",
	}

	ui.Say("Creating hard drive...")
	err := driver.VBoxManage(command...)
	if err != nil {
		ui.Error(fmt.Sprintf("Error creating hard drive: %s", err))
		return multistep.ActionHalt
	}

	// Set the path so that we can delete it later
	s.diskPath = path

	time.Sleep(15 * time.Second)
	return multistep.ActionContinue
}

func (s *stepCreateDisk) Cleanup(state map[string]interface{}) {
	if s.diskPath == "" {
		return
	}

	driver := state["driver"].(Driver)
	ui := state["ui"].(packer.Ui)

	ui.Say("Unregistering and deleting hard disk...")
	if err := driver.VBoxManage("closemedium", "disk", s.diskPath, "--delete"); err != nil {
		ui.Error(fmt.Sprintf("Error deleting hard drive: %s", err))
	}
}

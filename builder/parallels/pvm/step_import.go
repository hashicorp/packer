package pvm

import (
	"fmt"
	"github.com/mitchellh/multistep"
	parallelscommon "github.com/mitchellh/packer/builder/parallels/common"
	"github.com/mitchellh/packer/packer"
)

// This step imports an PVM VM into Parallels.
type StepImport struct {
	Name       string
	SourcePath string
	vmName     string
}

func (s *StepImport) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(parallelscommon.Driver)
	ui := state.Get("ui").(packer.Ui)
	config := state.Get("config").(*Config)

	ui.Say(fmt.Sprintf("Importing VM: %s", s.SourcePath))
	if err := driver.Import(s.Name, s.SourcePath, config.OutputDir); err != nil {
		err := fmt.Errorf("Error importing VM: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	s.vmName = s.Name
	state.Put("vmName", s.Name)
	return multistep.ActionContinue
}

func (s *StepImport) Cleanup(state multistep.StateBag) {

	if s.vmName == "" {
		return
	}

	driver := state.Get("driver").(parallelscommon.Driver)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Unregistering virtual machine...")
	if err := driver.Prlctl("unregister", s.vmName); err != nil {
		ui.Error(fmt.Sprintf("Error unregistering virtual machine: %s", err))
	}
}

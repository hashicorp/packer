package ovf

import (
	"fmt"
	vboxcommon "github.com/hashicorp/packer/builder/virtualbox/common"
	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
)

// This step imports an OVF VM into VirtualBox.
type StepImport struct {
	Name        string
	ImportFlags []string

	vmName string
}

func (s *StepImport) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(vboxcommon.Driver)
	ui := state.Get("ui").(packer.Ui)
	vmPath := state.Get("vm_path").(string)

	ui.Say(fmt.Sprintf("Importing VM: %s", vmPath))
	if err := driver.Import(s.Name, vmPath, s.ImportFlags); err != nil {
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

	driver := state.Get("driver").(vboxcommon.Driver)
	ui := state.Get("ui").(packer.Ui)
	config := state.Get("config").(*Config)

	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)
	if (config.KeepRegistered) && (!cancelled && !halted) {
		ui.Say("Keeping virtual machine registered with VirtualBox host (keep_registered = true)")
		return
	}

	ui.Say("Unregistering and deleting imported VM...")
	if err := driver.Delete(s.vmName); err != nil {
		ui.Error(fmt.Sprintf("Error deleting VM: %s", err))
	}
}

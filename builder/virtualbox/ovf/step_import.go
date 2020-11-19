package ovf

import (
	"context"
	"fmt"

	vboxcommon "github.com/hashicorp/packer/builder/virtualbox/common"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

// This step imports an OVF VM into VirtualBox.
type StepImport struct {
	Name           string
	ImportFlags    []string
	KeepRegistered bool

	vmName string
}

func (s *StepImport) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(vboxcommon.Driver)
	ui := state.Get("ui").(packersdk.Ui)
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
	ui := state.Get("ui").(packersdk.Ui)

	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)
	if (s.KeepRegistered) && (!cancelled && !halted) {
		ui.Say("Keeping virtual machine registered with VirtualBox host (keep_registered = true)")
		return
	}

	ui.Say("Deregistering and deleting imported VM...")
	if err := driver.Delete(s.vmName); err != nil {
		ui.Error(fmt.Sprintf("Error deleting VM: %s", err))
	}
}

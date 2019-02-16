package vm

import (
	"context"
	"fmt"

	vboxcommon "github.com/hashicorp/packer/builder/virtualbox/common"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type StepSetSnapshot struct {
	Name             string
	AttachSnapshot   string
	revertToSnapshot string
}

func (s *StepSetSnapshot) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(vboxcommon.Driver)
	if s.AttachSnapshot != "" {
		ui := state.Get("ui").(packer.Ui)
		hasSnapshots, err := driver.HasSnapshots(s.Name)
		if err != nil {
			err := fmt.Errorf("Error checking for snapshots VM: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		if !hasSnapshots {
			err := fmt.Errorf("Unable to attach snapshot on VM %s when no snapshots exist", s.Name)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		currentSnapshot, err := driver.GetCurrentSnapshot(s.Name)
		if err != nil {
			err := fmt.Errorf("Unable to get current snapshot for VM: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		ui.Say(fmt.Sprintf("Attaching snapshot %s on virtual machine %s", s.AttachSnapshot, s.Name))
		err = driver.SetSnapshot(s.Name, s.AttachSnapshot)
		if err != nil {
			err := fmt.Errorf("Unable to set snapshot for VM: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		s.revertToSnapshot = currentSnapshot
	}
	return multistep.ActionContinue
}

func (s *StepSetSnapshot) Cleanup(state multistep.StateBag) {
	driver := state.Get("driver").(vboxcommon.Driver)
	if s.revertToSnapshot != "" {
		ui := state.Get("ui").(packer.Ui)
		ui.Say(fmt.Sprintf("Reverting to snapshot %s on virtual machine %s", s.revertToSnapshot, s.Name))
		err := driver.SetSnapshot(s.Name, s.revertToSnapshot)
		if err != nil {
			err := fmt.Errorf("Unable to set snapshot for VM: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return
		}
	}
}

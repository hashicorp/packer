package vm

import (
	"context"
	"fmt"

	vboxcommon "github.com/hashicorp/packer/builder/virtualbox/common"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type StepSetSnapshot struct {
	Name             string
	AttachSnapshot   string
	KeepRegistered   bool
	revertToSnapshot string
}

func (s *StepSetSnapshot) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(vboxcommon.Driver)
	ui := state.Get("ui").(packersdk.Ui)
	snapshotTree, err := driver.LoadSnapshots(s.Name)
	if err != nil {
		err := fmt.Errorf("Error loading snapshots for VM: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if s.AttachSnapshot != "" {
		if nil == snapshotTree {
			err := fmt.Errorf("Unable to attach snapshot on VM %s when no snapshots exist", s.Name)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		currentSnapshot := snapshotTree.GetCurrentSnapshot()
		s.revertToSnapshot = currentSnapshot.UUID
		ui.Say(fmt.Sprintf("Attaching snapshot %s on virtual machine %s", s.AttachSnapshot, s.Name))
		candidateSnapshots := snapshotTree.GetSnapshotsByName(s.AttachSnapshot)
		if 0 >= len(candidateSnapshots) {
			err := fmt.Errorf("Snapshot %s not found on VM %s", s.AttachSnapshot, s.Name)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		} else if 1 > len(candidateSnapshots) {
			err := fmt.Errorf("More than one Snapshot %s found on VM %s", s.AttachSnapshot, s.Name)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		} else {
			err = driver.SetSnapshot(s.Name, candidateSnapshots[0])
			if err != nil {
				err := fmt.Errorf("Unable to set snapshot for VM: %s", err)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}
		}
	}
	return multistep.ActionContinue
}

func (s *StepSetSnapshot) Cleanup(state multistep.StateBag) {
	driver := state.Get("driver").(vboxcommon.Driver)
	if s.revertToSnapshot != "" {
		ui := state.Get("ui").(packersdk.Ui)
		if s.KeepRegistered {
			ui.Say("Keeping virtual machine state (keep_registered = true)")
			return
		} else {
			ui.Say(fmt.Sprintf("Reverting to snapshot %s on virtual machine %s", s.revertToSnapshot, s.Name))
			snapshotTree, err := driver.LoadSnapshots(s.Name)
			if err != nil {
				err := fmt.Errorf("error loading virtual machine %s snapshots: %v", s.Name, err)
				state.Put("error", err)
				ui.Error(err.Error())
				return
			}
			revertTo := snapshotTree.GetSnapshotByUUID(s.revertToSnapshot)
			if nil == revertTo {
				err := fmt.Errorf("Snapshot with UUID %s not found for VM %s", s.revertToSnapshot, s.Name)
				state.Put("error", err)
				ui.Error(err.Error())
				return
			}
			err = driver.SetSnapshot(s.Name, revertTo)
			if err != nil {
				err := fmt.Errorf("Unable to set snapshot for VM: %s", err)
				state.Put("error", err)
				ui.Error(err.Error())
				return
			}
		}
	}
}

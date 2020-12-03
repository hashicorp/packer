package vm

import (
	"context"
	"fmt"
	"log"

	vboxcommon "github.com/hashicorp/packer/builder/virtualbox/common"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type StepCreateSnapshot struct {
	Name           string
	TargetSnapshot string
}

func (s *StepCreateSnapshot) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(vboxcommon.Driver)
	ui := state.Get("ui").(packersdk.Ui)
	if s.TargetSnapshot != "" {
		running, err := driver.IsRunning(s.Name)
		if err != nil {
			err = fmt.Errorf("Failed to test if VM %s is still running: %s", s.Name, err)
		} else if running {
			err = fmt.Errorf("VM %s is still running. Unable to create snapshot %s", s.Name, s.TargetSnapshot)
		}
		if err != nil {
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		ui.Say(fmt.Sprintf("Creating snapshot %s on virtual machine %s", s.TargetSnapshot, s.Name))
		snapshotTree, err := driver.LoadSnapshots(s.Name)
		if err != nil {
			err = fmt.Errorf("Failed to load snapshots for VM %s: %s", s.Name, err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		// Remove any snapshot with the target's name, if present.
		if snapshotTree != nil {
			currentSnapshot := snapshotTree.GetCurrentSnapshot()
			targetSnapshot := currentSnapshot.GetChildWithName(s.TargetSnapshot)
			if nil != targetSnapshot {
				log.Printf("Deleting existing target snapshot %s", s.TargetSnapshot)
				err = driver.DeleteSnapshot(s.Name, targetSnapshot)
				if nil != err {
					err = fmt.Errorf("Unable to delete snapshot %s from VM %s: %s", s.TargetSnapshot, s.Name, err)
					state.Put("error", err)
					ui.Error(err.Error())
					return multistep.ActionHalt
				}
			}
		}

		err = driver.CreateSnapshot(s.Name, s.TargetSnapshot)
		if err != nil {
			err := fmt.Errorf("Error creating snaphot VM: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	} else {
		ui.Say("No target snapshot defined...")
	}

	return multistep.ActionContinue
}

func (s *StepCreateSnapshot) Cleanup(state multistep.StateBag) {}

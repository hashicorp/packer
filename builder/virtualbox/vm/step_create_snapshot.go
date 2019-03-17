package vm

import (
	"context"
	"fmt"
	"log"
	"time"

	vboxcommon "github.com/hashicorp/packer/builder/virtualbox/common"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type StepCreateSnapshot struct {
	Name           string
	TargetSnapshot string
}

func (s *StepCreateSnapshot) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(vboxcommon.Driver)
	ui := state.Get("ui").(packer.Ui)
	if s.TargetSnapshot != "" {
		time.Sleep(10 * time.Second) // Wait after the Vm has been shutdown, otherwise creating the snapshot might make the VM unstartable
		ui.Say(fmt.Sprintf("Creating snapshot %s on virtual machine %s", s.TargetSnapshot, s.Name))
		snapshotTree, err := driver.LoadSnapshots(s.Name)
		if err != nil {
			err = fmt.Errorf("Failed to load snapshots for VM %s: %s", s.Name, err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

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

func (s *StepCreateSnapshot) Cleanup(state multistep.StateBag) {
	/*
		driver := state.Get("driver").(vboxcommon.Driver)
		if s.TargetSnapshot != "" {
			ui := state.Get("ui").(packer.Ui)
			ui.Say(fmt.Sprintf("Deleting snapshot %s on virtual machine %s", s.TargetSnapshot, s.Name))
			err := driver.DeleteSnapshot(s.Name, s.TargetSnapshot)
			if err != nil {
				err := fmt.Errorf("Error cleaning up created snaphot VM: %s", err)
				state.Put("error", err)
				ui.Error(err.Error())
				return
			}
		}
	*/
}

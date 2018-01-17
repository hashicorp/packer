package classic

import (
	"fmt"

	"github.com/hashicorp/go-oracle-terraform/compute"
	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
)

type stepSnapshot struct{}

func (s *stepSnapshot) Run(state multistep.StateBag) multistep.StepAction {
	// get variables from state
	ui := state.Get("ui").(packer.Ui)
	ui.Say("Creating Snapshot...")
	config := state.Get("config").(Config)
	client := state.Get("client").(*compute.ComputeClient)
	instanceID := state.Get("instance_id").(string)

	// get instances client
	snapshotClient := client.Snapshots()

	// Instances Input
	snapshotInput := &compute.CreateSnapshotInput{
		Instance:     instanceID,
		MachineImage: config.ImageName,
	}

	snap, err := snapshotClient.CreateSnapshot(snapshotInput)
	if err != nil {
		err = fmt.Errorf("Problem creating snapshot: %s", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

	ui.Say(fmt.Sprintf("Created snapshot (%s).", snap.Name))
	return multistep.ActionContinue
}

func (s *stepSnapshot) Cleanup(state multistep.StateBag) {
	// Nothing to do
}

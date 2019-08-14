package vultr

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/vultr/govultr"
)

type stepCreateSnapshot struct {
	client *govultr.Client
}

func (s *stepCreateSnapshot) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	c := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)
	server := state.Get("server").(*govultr.Server)

	s.client = govultr.NewClient(nil, c.APIKey)

	snapshot, err := s.client.Snapshot.Create(ctx, server.InstanceID, c.Description)
	if err != nil {
		err := fmt.Errorf("Error creating snapshot: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say(fmt.Sprintf("Waiting %ds for snapshot %s to complete...",
		int(c.stateTimeout/time.Second), snapshot.SnapshotID))

	err = waitForSnapshotState("complete", snapshot.SnapshotID, s.client, c.stateTimeout)
	if err != nil {
		err := fmt.Errorf("Error waiting for snapshot: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("snapshot", snapshot)
	return multistep.ActionContinue
}

func (s *stepCreateSnapshot) Cleanup(state multistep.StateBag) {
}

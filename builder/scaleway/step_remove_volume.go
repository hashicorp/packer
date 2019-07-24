package scaleway

import (
	"context"
	"fmt"

	"github.com/scaleway/scaleway-cli/pkg/api"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepRemoveVolume struct{}

func (s *stepRemoveVolume) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	// nothing to do ... only cleanup interests us
	return multistep.ActionContinue
}

func (s *stepRemoveVolume) Cleanup(state multistep.StateBag) {
	if _, ok := state.GetOk("snapshot_name"); !ok {
		// volume will be detached from server only after snapshotting ... so we don't
		// need to remove volume before snapshot step.
		return
	}

	client := state.Get("client").(*api.ScalewayAPI)
	ui := state.Get("ui").(packer.Ui)
	c := state.Get("config").(*Config)
	volumeID := state.Get("root_volume_id").(string)

	if !c.RemoveVolume {
		return
	}

	ui.Say("Removing Volume ...")

	err := client.DeleteVolume(volumeID)
	if err != nil {
		err := fmt.Errorf("Error removing volume: %s", err)
		state.Put("error", err)
		ui.Error(fmt.Sprintf("Error removing volume: %s. (Ignored)", err))
	}
}

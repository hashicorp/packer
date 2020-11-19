package scaleway

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
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

	instanceAPI := instance.NewAPI(state.Get("client").(*scw.Client))
	ui := state.Get("ui").(packersdk.Ui)
	c := state.Get("config").(*Config)
	volumeID := state.Get("root_volume_id").(string)

	if !c.RemoveVolume {
		return
	}

	ui.Say("Removing Volume ...")

	err := instanceAPI.DeleteVolume(&instance.DeleteVolumeRequest{
		VolumeID: volumeID,
	})
	if err != nil {
		err := fmt.Errorf("Error removing volume: %s", err)
		state.Put("error", err)
		ui.Error(fmt.Sprintf("Error removing volume: %s. (Ignored)", err))
	}
}

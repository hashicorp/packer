package hcloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

type stepCreateSnapshot struct{}

func (s *stepCreateSnapshot) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("hcloudClient").(*hcloud.Client)
	ui := state.Get("ui").(packer.Ui)
	c := state.Get("config").(*Config)
	serverID := state.Get("server_id").(int)

	ui.Say("Creating snapshot ...")
	ui.Say("This can take some time")
	result, _, err := client.Server.CreateImage(context.TODO(), &hcloud.Server{ID: serverID}, &hcloud.ServerCreateImageOpts{
		Type:        hcloud.ImageTypeSnapshot,
		Description: hcloud.String(c.SnapshotName),
	})
	if err != nil {
		err := fmt.Errorf("Error creating snapshot: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	state.Put("snapshot_id", result.Image.ID)
	state.Put("snapshot_name", c.SnapshotName)
	_, errCh := client.Action.WatchProgress(context.TODO(), result.Action)
	for {
		select {
		case err1 := <-errCh:
			if err1 == nil {
				return multistep.ActionContinue
			} else {
				err := fmt.Errorf("Error creating snapshot: %s", err)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}

		}
	}
}

func (s *stepCreateSnapshot) Cleanup(state multistep.StateBag) {
	// no cleanup
}

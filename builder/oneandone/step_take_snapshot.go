package oneandone

import (
	"context"

	"github.com/1and1/oneandone-cloudserver-sdk-go"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type stepTakeSnapshot struct{}

func (s *stepTakeSnapshot) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	c := state.Get("config").(*Config)

	ui.Say("Creating Snapshot...")

	token := oneandone.SetToken(c.Token)
	api := oneandone.New(token, c.Url)

	serverId := state.Get("server_id").(string)

	req := oneandone.ImageConfig{
		Name:        c.SnapshotName,
		Description: "Packer image",
		ServerId:    serverId,
		Frequency:   "WEEKLY",
		NumImages:   1,
	}

	img_id, img, err := api.CreateImage(&req)

	if err != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	err = api.WaitForState(img, "ENABLED", 10, c.Retries)

	if err != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("snapshot_id", img_id)
	state.Put("snapshot_name", img.Name)
	return multistep.ActionContinue
}

func (s *stepTakeSnapshot) Cleanup(state multistep.StateBag) {
}

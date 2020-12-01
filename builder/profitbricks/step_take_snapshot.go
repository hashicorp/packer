package profitbricks

import (
	"context"
	"encoding/json"
	"time"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/profitbricks/profitbricks-sdk-go"
)

type stepTakeSnapshot struct{}

func (s *stepTakeSnapshot) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	c := state.Get("config").(*Config)

	ui.Say("Creating ProfitBricks snapshot...")

	profitbricks.SetAuth(c.PBUsername, c.PBPassword)

	dcId := state.Get("datacenter_id").(string)
	volumeId := state.Get("volume_id").(string)

	snapshot := profitbricks.CreateSnapshot(dcId, volumeId, c.SnapshotName, "")

	state.Put("snapshotname", c.SnapshotName)

	if snapshot.StatusCode > 299 {
		var restError RestError
		if err := json.Unmarshal([]byte(snapshot.Response), &restError); err != nil {
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		if len(restError.Messages) > 0 {
			ui.Error(restError.Messages[0].Message)
		} else {
			ui.Error(snapshot.Response)
		}

		return multistep.ActionHalt
	}

	s.waitTillProvisioned(snapshot.Headers.Get("Location"), *c)

	return multistep.ActionContinue
}

func (s *stepTakeSnapshot) Cleanup(state multistep.StateBag) {
}

func (d *stepTakeSnapshot) waitTillProvisioned(path string, config Config) {
	d.setPB(config.PBUsername, config.PBPassword, config.PBUrl)
	waitCount := 50
	if config.Retries > 0 {
		waitCount = config.Retries
	}
	for i := 0; i < waitCount; i++ {
		request := profitbricks.GetRequestStatus(path)
		if request.Metadata.Status == "DONE" {
			break
		}
		time.Sleep(10 * time.Second)
		i++
	}
}

func (d *stepTakeSnapshot) setPB(username string, password string, url string) {
	profitbricks.SetAuth(username, password)
	profitbricks.SetEndpoint(url)
}

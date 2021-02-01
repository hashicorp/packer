package profitbricks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
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
	serverId := state.Get("instance_id").(string)

	comm, _ := state.Get("communicator").(packersdk.Communicator)
	if comm == nil {
		ui.Error("no communicator found")
		return multistep.ActionHalt
	}

	/* sync fs changes from the provisioning step */
	os, err := s.getOs(dcId, serverId)
	if err != nil {
		ui.Error(fmt.Sprintf("an error occurred while getting the server os: %s", err.Error()))
		return multistep.ActionHalt
	}
	ui.Say(fmt.Sprintf("Server OS is %s", os))

	switch strings.ToLower(os) {
	case "linux":
		ui.Say("syncing file system changes")
		if err := s.syncFs(ctx, comm); err != nil {
			ui.Error(fmt.Sprintf("error syncing fs changes: %s", err.Error()))
			return multistep.ActionHalt
		}
	}

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

	ui.Say(fmt.Sprintf("Creating a snapshot for %s/volumes/%s", dcId, volumeId))

	err = s.waitForRequest(snapshot.Headers.Get("Location"), *c, ui)
	if err != nil {
		ui.Error(fmt.Sprintf("An error occurred while waiting for the request to be done: %s", err.Error()))
		return multistep.ActionHalt
	}

	err = s.waitTillSnapshotAvailable(snapshot.Id, *c, ui)
	if err != nil {
		ui.Error(fmt.Sprintf("An error occurred while waiting for the snapshot to be created: %s", err.Error()))
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepTakeSnapshot) Cleanup(_ multistep.StateBag) {
}

func (s *stepTakeSnapshot) waitForRequest(path string, config Config, ui packersdk.Ui) error {

	ui.Say(fmt.Sprintf("Watching request %s", path))
	s.setPB(config.PBUsername, config.PBPassword, config.PBUrl)
	waitCount := 50
	var waitInterval = 10 * time.Second
	if config.Retries > 0 {
		waitCount = config.Retries
	}
	done := false
	for i := 0; i < waitCount; i++ {
		request := profitbricks.GetRequestStatus(path)
		ui.Say(fmt.Sprintf("request status = %s", request.Metadata.Status))
		if request.Metadata.Status == "DONE" {
			done = true
			break
		}
		if request.Metadata.Status == "FAILED" {
			return fmt.Errorf("Request failed: %s", request.Response)
		}
		time.Sleep(waitInterval)
		i++
	}

	if done == false {
		return fmt.Errorf("request not fulfilled after waiting %d seconds",
			int64(waitCount)*int64(waitInterval)/int64(time.Second))
	}
	return nil
}

func (s *stepTakeSnapshot) waitTillSnapshotAvailable(id string, config Config, ui packersdk.Ui) error {
	s.setPB(config.PBUsername, config.PBPassword, config.PBUrl)
	waitCount := 50
	var waitInterval = 10 * time.Second
	if config.Retries > 0 {
		waitCount = config.Retries
	}
	done := false
	ui.Say(fmt.Sprintf("waiting for snapshot %s to become available", id))
	for i := 0; i < waitCount; i++ {
		snap := profitbricks.GetSnapshot(id)
		ui.Say(fmt.Sprintf("snapshot status = %s", snap.Metadata.State))
		if snap.StatusCode != 200 {
			return fmt.Errorf("%s", snap.Response)
		}
		if snap.Metadata.State == "AVAILABLE" {
			done = true
			break
		}
		time.Sleep(waitInterval)
		i++
		ui.Say(fmt.Sprintf("... still waiting, %d seconds have passed", int64(waitInterval)*int64(i)))
	}

	if done == false {
		return fmt.Errorf("snapshot not created after waiting %d seconds",
			int64(waitCount)*int64(waitInterval)/int64(time.Second))
	}

	ui.Say("snapshot created")
	return nil
}

func (s *stepTakeSnapshot) syncFs(ctx context.Context, comm packersdk.Communicator) error {
	cmd := &packersdk.RemoteCmd{
		Command: "sync",
	}
	if err := comm.Start(ctx, cmd); err != nil {
		return err
	}
	if cmd.Wait() != 0 {
		return fmt.Errorf("sync command exited with code %d", cmd.ExitStatus())
	}
	return nil
}

func (s *stepTakeSnapshot) getOs(dcId string, serverId string) (string, error) {
	server := profitbricks.GetServer(dcId, serverId)
	if server.StatusCode != 200 {
		return "", errors.New(server.Response)
	}

	if server.Properties.BootVolume == nil {
		return "", errors.New("no boot volume found on server")
	}

	volumeId := server.Properties.BootVolume.Id
	volume := profitbricks.GetVolume(dcId, volumeId)
	if volume.StatusCode != 200 {
		return "", errors.New(volume.Response)
	}

	return volume.Properties.LicenceType, nil
}

func (s *stepTakeSnapshot) setPB(username string, password string, url string) {
	profitbricks.SetAuth(username, password)
	profitbricks.SetEndpoint(url)
}

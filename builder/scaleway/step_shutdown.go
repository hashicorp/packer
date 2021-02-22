package scaleway

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

type stepShutdown struct{}

func (s *stepShutdown) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	instanceAPI := instance.NewAPI(state.Get("client").(*scw.Client))
	ui := state.Get("ui").(packersdk.Ui)
	serverID := state.Get("server_id").(string)

	ui.Say("Shutting down server...")

	_, err := instanceAPI.ServerAction(&instance.ServerActionRequest{
		Action:   instance.ServerActionPoweroff,
		ServerID: serverID,
	}, scw.WithContext(ctx))
	if err != nil {
		err := fmt.Errorf("Error stopping server: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	waitRequest := &instance.WaitForServerRequest{
		ServerID: serverID,
	}
	c := state.Get("config").(*Config)
	timeout := c.ShutdownTimeout
	duration, err := time.ParseDuration(timeout)
	if err != nil {
		err := fmt.Errorf("error: %s could not parse string %s as a duration", err, timeout)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	if timeout != "" {
		waitRequest.Timeout = scw.TimeDurationPtr(duration)
	}

	instanceResp, err := instanceAPI.WaitForServer(waitRequest)
	if err != nil {
		err := fmt.Errorf("Error shutting down server: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if instanceResp.State != instance.ServerStateStopped {
		err := fmt.Errorf("Server is in state %s instead of stopped", instanceResp.State.String())
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepShutdown) Cleanup(state multistep.StateBag) {
	// no cleanup
}

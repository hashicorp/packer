package vultr

import (
	"context"
	"github.com/vultr/govultr"
	"time"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// shutdown delays
const (
	ShutdownDelaySec = 10
)

type stepShutdown struct {
	client *govultr.Client
}

func (s *stepShutdown) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	c := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Performing graceful shutdown...")
	time.Sleep(ShutdownDelaySec * time.Second)
	id := state.Get("server_id").(string)
	ui.Say("Performing graceful shutdown...")

	err := s.client.Server.Halt(context.Background(), id)

	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	err = waitForServerState("active", "stopped", id, s.client, c.stateTimeout)
	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepShutdown) Cleanup(state multistep.StateBag) {
}

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
	ui := state.Get("ui").(packer.Ui)
	ui.Say("Performing graceful shutdown...")
	time.Sleep(ShutdownDelaySec * time.Second)

	comm := state.Get("communicator").(packer.Communicator)

	cmd := &packer.RemoteCmd{
		Command: "shutdown -P now",
	}

	if err := comm.Start(ctx, cmd); err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	cmd.Wait()

	if cmd.ExitStatus() == packer.CmdDisconnect {
		ui.Say("Server successfully shutdown")
		time.Sleep(ShutdownDelaySec * time.Second)
	} else {
		ui.Say("Sleeping to ensure that server is shut down...")
	}

	return multistep.ActionContinue
}

func (s *stepShutdown) Cleanup(state multistep.StateBag) {
}

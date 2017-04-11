package scaleway

import (
	"fmt"

	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
	"github.com/scaleway/scaleway-cli/pkg/api"
)

type stepTerminate struct{}

func (s *stepTerminate) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*api.ScalewayAPI)
	ui := state.Get("ui").(packer.Ui)
	serverID := state.Get("server_id").(string)

	ui.Say("Terminating server...")

	err := client.DeleteServerForce(serverID)

	if err != nil {
		err := fmt.Errorf("Error terminating server: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepTerminate) Cleanup(state multistep.StateBag) {
	// no cleanup
}

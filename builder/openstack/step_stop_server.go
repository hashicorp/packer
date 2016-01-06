package openstack

import (
	"fmt"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"github.com/rackspace/gophercloud/openstack/compute/v2/extensions/startstop"
	"github.com/rackspace/gophercloud/openstack/compute/v2/servers"
)

type StepStopServer struct{}

func (s *StepStopServer) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	config := state.Get("config").(Config)
	extensions := state.Get("extensions").(map[string]struct{})
	server := state.Get("server").(*servers.Server)

	// Verify we have the extension
	if _, ok := extensions["os-server-start-stop"]; !ok {
		ui.Say("OpenStack cluster doesn't support stop, skipping...")
		return multistep.ActionContinue
	}

	// We need the v2 compute client
	client, err := config.computeV2Client()
	if err != nil {
		err = fmt.Errorf("Error initializing compute client: %s", err)
		state.Put("error", err)
		return multistep.ActionHalt
	}

	ui.Say(fmt.Sprintf("Stopping server: %s ...", server.ID))
	if err := startstop.Stop(client, server.ID).ExtractErr(); err != nil {
		err = fmt.Errorf("Error stopping server: %s", err)
		state.Put("error", err)
		return multistep.ActionHalt
	}

	ui.Message(fmt.Sprintf("Waiting for server to stop: %s ...", server.ID))
	stateChange := StateChangeConf{
		Pending:   []string{"ACTIVE"},
		Target:    []string{"SHUTOFF", "STOPPED"},
		Refresh:   ServerStateRefreshFunc(client, server),
		StepState: state,
	}
	if _, err := WaitForState(&stateChange); err != nil {
		err := fmt.Errorf("Error waiting for server (%s) to stop: %s", server.ID, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepStopServer) Cleanup(state multistep.StateBag) {}

package openstack

import (
	"fmt"
	"log"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"github.com/rackspace/gophercloud/openstack/compute/v2/extensions/keypairs"
	"github.com/rackspace/gophercloud/openstack/compute/v2/servers"
)

type StepRunSourceServer struct {
	Flavor         string
	Name           string
	SourceImage    string
	SecurityGroups []string
	Networks       []string

	server *servers.Server
}

func (s *StepRunSourceServer) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(Config)
	keyName := state.Get("keyPair").(string)
	ui := state.Get("ui").(packer.Ui)

	// We need the v2 compute client
	computeClient, err := config.computeV2Client()
	if err != nil {
		err = fmt.Errorf("Error initializing compute client: %s", err)
		state.Put("error", err)
		return multistep.ActionHalt
	}

	networks := make([]servers.Network, len(s.Networks))
	for i, networkUuid := range s.Networks {
		networks[i].UUID = networkUuid
	}

	s.server, err = servers.Create(computeClient, keypairs.CreateOptsExt{
		CreateOptsBuilder: servers.CreateOpts{
			Name:           s.Name,
			ImageRef:       s.SourceImage,
			FlavorName:     s.Flavor,
			SecurityGroups: s.SecurityGroups,
			Networks:       networks,
		},

		KeyName: keyName,
	}).Extract()
	if err != nil {
		err := fmt.Errorf("Error launching source server: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	log.Printf("server id: %s", s.server.ID)

	ui.Say(fmt.Sprintf("Waiting for server (%s) to become ready...", s.server.ID))
	stateChange := StateChangeConf{
		Pending:   []string{"BUILD"},
		Target:    "ACTIVE",
		Refresh:   ServerStateRefreshFunc(computeClient, s.server),
		StepState: state,
	}
	latestServer, err := WaitForState(&stateChange)
	if err != nil {
		err := fmt.Errorf("Error waiting for server (%s) to become ready: %s", s.server.ID, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	s.server = latestServer.(*servers.Server)
	state.Put("server", s.server)

	return multistep.ActionContinue
}

func (s *StepRunSourceServer) Cleanup(state multistep.StateBag) {
	if s.server == nil {
		return
	}

	config := state.Get("config").(Config)
	ui := state.Get("ui").(packer.Ui)

	// We need the v2 compute client
	computeClient, err := config.computeV2Client()
	if err != nil {
		ui.Error(fmt.Sprintf("Error terminating server, may still be around: %s", err))
		return
	}

	ui.Say("Terminating the source server...")
	if err := servers.Delete(computeClient, s.server.ID).ExtractErr(); err != nil {
		ui.Error(fmt.Sprintf("Error terminating server, may still be around: %s", err))
		return
	}

	stateChange := StateChangeConf{
		Pending: []string{"ACTIVE", "BUILD", "REBUILD", "SUSPENDED"},
		Refresh: ServerStateRefreshFunc(computeClient, s.server),
		Target:  "DELETED",
	}

	WaitForState(&stateChange)
}

package openstack

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"github.com/rackspace/gophercloud/openstack/compute/v2/extensions/keypairs"
	"github.com/rackspace/gophercloud/openstack/compute/v2/servers"
)

type StepRunSourceServer struct {
	Name             string
	SourceImage      string
	SecurityGroups   []string
	Networks         []string
	AvailabilityZone string
	UserData         string
	UserDataFile     string

	server *servers.Server
}

func (s *StepRunSourceServer) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(Config)
	flavor := state.Get("flavor_id").(string)
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

	userData := []byte(s.UserData)
	if s.UserDataFile != "" {
		userData, err = ioutil.ReadFile(s.UserDataFile)
		if err != nil {
			err = fmt.Errorf("Error reading user data file: %s", err)
			state.Put("error", err)
			return multistep.ActionHalt
		}
	}

	ui.Say("Launching server...")
	s.server, err = servers.Create(computeClient, keypairs.CreateOptsExt{
		CreateOptsBuilder: servers.CreateOpts{
			Name:             s.Name,
			ImageRef:         s.SourceImage,
			FlavorRef:        flavor,
			SecurityGroups:   s.SecurityGroups,
			Networks:         networks,
			AvailabilityZone: s.AvailabilityZone,
			UserData:         userData,
		},

		KeyName: keyName,
	}).Extract()
	if err != nil {
		err := fmt.Errorf("Error launching source server: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Message(fmt.Sprintf("Server ID: %s", s.server.ID))
	log.Printf("server id: %s", s.server.ID)

	ui.Say("Waiting for server to become ready...")
	stateChange := StateChangeConf{
		Pending:   []string{"BUILD"},
		Target:    []string{"ACTIVE"},
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
		Pending: []string{"ACTIVE", "BUILD", "REBUILD", "SUSPENDED", "SHUTOFF", "STOPPED"},
		Refresh: ServerStateRefreshFunc(computeClient, s.server),
		Target:  []string{"DELETED"},
	}

	WaitForState(&stateChange)
}

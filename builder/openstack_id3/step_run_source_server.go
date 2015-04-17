package openstack_id3

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"

	"github.com/rackspace/gophercloud"
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

	compute_client := state.Get("compute_client").(*gophercloud.ServiceClient)
	keyName := state.Get("keyPair").(string)
	ui := state.Get("ui").(packer.Ui)

	// XXX - validate image and flavor is available

	securityGroups := make([]string, len(s.SecurityGroups))
	for i, groupName := range s.SecurityGroups {
		securityGroups[i] = groupName
	}

	networkList := make([]servers.Network, len(s.Networks))
	for i, networkUuid := range s.Networks {
		networkList[i] = servers.Network{UUID: networkUuid}
	}

	create_opts := servers.CreateOpts{
		Name:           s.Name,
		ImageRef:       s.SourceImage,
		FlavorRef:      s.Flavor,
		SecurityGroups: securityGroups,
		Networks:       networkList,
	}

	var err error
	s.server, err = servers.Create(compute_client, keypairs.CreateOptsExt{create_opts, keyName}).Extract()
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
		Refresh:   ServerStateRefreshFunc(compute_client, s.server),
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

	compute_client := state.Get("compute_client").(*gophercloud.ServiceClient)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Terminating the source server...")

	err := servers.Delete(compute_client, s.server.ID).ExtractErr()
	if err != nil {
		ui.Error(fmt.Sprintf("Error terminating server, may still be around: %s", err))
		return
	}

	stateChange := StateChangeConf{
		Pending: []string{"ACTIVE", "BUILD", "REBUILD", "SUSPENDED"},
		Refresh: ServerStateRefreshFunc(compute_client, s.server),
		Target:  "DELETED",
	}

	WaitForState(&stateChange)
}

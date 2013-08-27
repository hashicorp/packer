package openstack

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"github.com/rackspace/gophercloud"
	"log"
)

type StepRunSourceServer struct {
	Flavor      string
	Name        string
	SourceImage string

	server *gophercloud.Server
}

func (s *StepRunSourceServer) Run(state map[string]interface{}) multistep.StepAction {
	accessor := state["accessor"].(*gophercloud.Access)
	api := state["api"].(*gophercloud.ApiCriteria)
	keyName := state["keyPair"].(string)
	ui := state["ui"].(packer.Ui)

	csp, err := gophercloud.ServersApi(accessor, *api)
	if err != nil {
		err := fmt.Errorf("Error connecting to api: %s", err)
		state["error"] = err
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// XXX - validate image and flavor is available

	server := gophercloud.NewServer{
		Name:        s.Name,
		ImageRef:    s.SourceImage,
		FlavorRef:   s.Flavor,
		KeyPairName: keyName,
	}

	serverResp, err := csp.CreateServer(server)
	if err != nil {
		err := fmt.Errorf("Error launching source server: %s", err)
		state["error"] = err
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	s.server, err = csp.ServerById(serverResp.Id)
	log.Printf("server id: %s", s.server.Id)

	ui.Say(fmt.Sprintf("Waiting for server (%s) to become ready...", s.server.Id))
	stateChange := StateChangeConf{
		Accessor:  accessor,
		Api:       api,
		Pending:   []string{"BUILD"},
		Target:    "ACTIVE",
		Refresh:   ServerStateRefreshFunc(accessor, api, s.server),
		StepState: state,
	}
	latestServer, err := WaitForState(&stateChange)
	if err != nil {
		err := fmt.Errorf("Error waiting for server (%s) to become ready: %s", s.server.Id, err)
		state["error"] = err
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	s.server = latestServer.(*gophercloud.Server)
	state["server"] = s.server

	return multistep.ActionContinue
}

func (s *StepRunSourceServer) Cleanup(state map[string]interface{}) {
	if s.server == nil {
		return
	}

	accessor := state["accessor"].(*gophercloud.Access)
	api := state["api"].(*gophercloud.ApiCriteria)
	ui := state["ui"].(packer.Ui)

	csp, err := gophercloud.ServersApi(accessor, *api)
	if err != nil {
		err := fmt.Errorf("Error connecting to api: %s", err)
		state["error"] = err
		ui.Error(err.Error())
		return
	}

	ui.Say("Terminating the source server...")
	if err := csp.DeleteServerById(s.server.Id); err != nil {
		ui.Error(fmt.Sprintf("Error terminating server, may still be around: %s", err))
		return
	}

	stateChange := StateChangeConf{
		Accessor: accessor,
		Api:      api,
		Pending:  []string{"ACTIVE", "BUILD", "REBUILD", "SUSPENDED"},
		Refresh:  ServerStateRefreshFunc(accessor, api, s.server),
		Target:   "DELETED",
	}

	WaitForState(&stateChange)
}

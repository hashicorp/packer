package openstack

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"github.com/rackspace/gophercloud"
	"log"
)

type StepRunSourceServer struct {
	Flavor         string
	Name           string
	SourceImage    string
	SecurityGroups []string

	server *gophercloud.Server
}

func (s *StepRunSourceServer) Run(state multistep.StateBag) multistep.StepAction {
	csp := state.Get("csp").(gophercloud.CloudServersProvider)
	keyName := state.Get("keyPair").(string)
	ui := state.Get("ui").(packer.Ui)

	// XXX - validate image and flavor is available

	securityGroups := make([]map[string]interface{}, len(s.SecurityGroups))
	for i, groupName := range s.SecurityGroups {
		securityGroups[i] = make(map[string]interface{})
		securityGroups[i]["name"] = groupName
	}

	server := gophercloud.NewServer{
		Name:          s.Name,
		ImageRef:      s.SourceImage,
		FlavorRef:     s.Flavor,
		KeyPairName:   keyName,
		SecurityGroup: securityGroups,
	}

	serverResp, err := csp.CreateServer(server)
	if err != nil {
		err := fmt.Errorf("Error launching source server: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	s.server, err = csp.ServerById(serverResp.Id)
	log.Printf("server id: %s", s.server.Id)

	ui.Say(fmt.Sprintf("Waiting for server (%s) to become ready...", s.server.Id))
	stateChange := StateChangeConf{
		Pending:   []string{"BUILD"},
		Target:    "ACTIVE",
		Refresh:   ServerStateRefreshFunc(csp, s.server),
		StepState: state,
	}
	latestServer, err := WaitForState(&stateChange)
	if err != nil {
		err := fmt.Errorf("Error waiting for server (%s) to become ready: %s", s.server.Id, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	s.server = latestServer.(*gophercloud.Server)
	state.Put("server", s.server)

	return multistep.ActionContinue
}

func (s *StepRunSourceServer) Cleanup(state multistep.StateBag) {
	if s.server == nil {
		return
	}

	csp := state.Get("csp").(gophercloud.CloudServersProvider)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Terminating the source server...")
	if err := csp.DeleteServerById(s.server.Id); err != nil {
		ui.Error(fmt.Sprintf("Error terminating server, may still be around: %s", err))
		return
	}

	stateChange := StateChangeConf{
		Pending: []string{"ACTIVE", "BUILD", "REBUILD", "SUSPENDED"},
		Refresh: ServerStateRefreshFunc(csp, s.server),
		Target:  "DELETED",
	}

	WaitForState(&stateChange)
}

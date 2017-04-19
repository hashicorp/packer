package scaleway

import (
	"fmt"
	"strings"

	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
	"github.com/scaleway/scaleway-cli/pkg/api"
)

type stepCreateServer struct {
	serverID string
}

func (s *stepCreateServer) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*api.ScalewayAPI)
	ui := state.Get("ui").(packer.Ui)
	c := state.Get("config").(Config)
	sshPubKey := state.Get("ssh_pubkey").(string)
	tags := []string{}

	ui.Say("Creating server...")

	if sshPubKey != "" {
		tags = []string{fmt.Sprintf("AUTHORIZED_KEY=%s", strings.TrimSpace(sshPubKey))}
	}

	server, err := client.PostServer(api.ScalewayServerDefinition{
		Name:           c.ServerName,
		Image:          &c.Image,
		Organization:   c.Organization,
		CommercialType: c.CommercialType,
		Tags:           tags,
	})

	err = client.PostServerAction(server, "poweron")

	if err != nil {
		err := fmt.Errorf("Error creating server: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	s.serverID = server

	state.Put("server_id", server)

	return multistep.ActionContinue
}

func (s *stepCreateServer) Cleanup(state multistep.StateBag) {
	if s.serverID != "" {
		return
	}

	client := state.Get("client").(*api.ScalewayAPI)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Destroying server...")
	err := client.PostServerAction(s.serverID, "terminate")
	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error destroying server. Please destroy it manually: %s", err))
	}
}

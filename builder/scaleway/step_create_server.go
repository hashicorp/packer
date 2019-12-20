package scaleway

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/scaleway/scaleway-cli/pkg/api"
)

type stepCreateServer struct {
	serverID string
}

func (s *stepCreateServer) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*api.ScalewayAPI)
	ui := state.Get("ui").(packer.Ui)
	c := state.Get("config").(*Config)
	tags := []string{}
	var bootscript *string

	ui.Say("Creating server...")

	if c.Bootscript != "" {
		bootscript = &c.Bootscript
	}

	if c.Comm.SSHPublicKey != nil {
		tags = []string{fmt.Sprintf("AUTHORIZED_KEY=%s", strings.Replace(strings.TrimSpace(string(c.Comm.SSHPublicKey)), " ", "_", -1))}
	}

	server, err := client.PostServer(api.ScalewayServerDefinition{
		Name:           c.ServerName,
		Image:          &c.Image,
		Organization:   c.Organization,
		CommercialType: c.CommercialType,
		Tags:           tags,
		Bootscript:     bootscript,
		BootType:       c.BootType,
	})

	if err != nil {
		err := fmt.Errorf("Error creating server: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	err = client.PostServerAction(server, "poweron")

	if err != nil {
		err := fmt.Errorf("Error starting server: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	s.serverID = server

	state.Put("server_id", server)
	// instance_id is the generic term used so that users can have access to the
	// instance id inside of the provisioners, used in step_provision.
	state.Put("instance_id", s.serverID)

	return multistep.ActionContinue
}

func (s *stepCreateServer) Cleanup(state multistep.StateBag) {
	if s.serverID == "" {
		return
	}

	client := state.Get("client").(*api.ScalewayAPI)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Destroying server...")

	err := client.DeleteServerForce(s.serverID)

	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error destroying server. Please destroy it manually: %s", err))
	}

}

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

func (s *stepCreateServer) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*api.ScalewayAPI)
	ui := state.Get("ui").(packer.Ui)
	c := state.Get("config").(*Config)
	env := ""
	var bootscript *string

	ui.Say("Creating server...")

	if c.Bootscript != "" {
		bootscript = &c.Bootscript
	}

	if c.Comm.SSHPublicKey != nil {
		env = fmt.Sprintf("AUTHORIZED_KEY=%s", strings.Replace(strings.TrimSpace(string(c.Comm.SSHPublicKey)), " ", "_", -1))
	}

	serverconfig := api.ConfigCreateServer{
		ImageName:         c.Image,
		Name:              c.ServerName,
		IP:                "",
		DynamicIPRequired: true,
		CommercialType:    c.CommercialType,
		Env:               env,
		Bootscript:        *bootscript,
		BootType:          c.BootType,
	}

	server, err := api.CreateServer(client, &serverconfig)

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

package scaleway

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

type stepCreateServer struct {
	serverID string
}

func (s *stepCreateServer) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	instanceAPI := instance.NewAPI(state.Get("client").(*scw.Client))
	ui := state.Get("ui").(packersdk.Ui)
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

	bootType := instance.BootType(c.BootType)

	createServerResp, err := instanceAPI.CreateServer(&instance.CreateServerRequest{
		BootType:       &bootType,
		Bootscript:     bootscript,
		CommercialType: c.CommercialType,
		Name:           c.ServerName,
		Image:          c.Image,
		Tags:           tags,
	}, scw.WithContext(ctx))
	if err != nil {
		err := fmt.Errorf("Error creating server: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	_, err = instanceAPI.ServerAction(&instance.ServerActionRequest{
		Action:   instance.ServerActionPoweron,
		ServerID: createServerResp.Server.ID,
	}, scw.WithContext(ctx))
	if err != nil {
		err := fmt.Errorf("Error starting server: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	s.serverID = createServerResp.Server.ID

	state.Put("server_id", createServerResp.Server.ID)
	// instance_id is the generic term used so that users can have access to the
	// instance id inside of the provisioners, used in step_provision.
	state.Put("instance_id", s.serverID)

	return multistep.ActionContinue
}

func (s *stepCreateServer) Cleanup(state multistep.StateBag) {
	if s.serverID == "" {
		return
	}

	instanceAPI := instance.NewAPI(state.Get("client").(*scw.Client))
	ui := state.Get("ui").(packersdk.Ui)

	ui.Say("Destroying server...")

	err := instanceAPI.DeleteServer(&instance.DeleteServerRequest{
		ServerID: s.serverID,
	})
	if err != nil {
		_, err = instanceAPI.ServerAction(&instance.ServerActionRequest{
			Action:   instance.ServerActionTerminate,
			ServerID: s.serverID,
		})
		if err != nil {
			ui.Error(fmt.Sprintf(
				"Error destroying server. Please destroy it manually: %s", err))
		}
	}
}

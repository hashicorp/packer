package vultr

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/vultr/govultr"
)

type stepCreateServer struct {
	client *govultr.Client
}

func (s *stepCreateServer) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	c := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Creating Vultr instance...")

	tempKey := state.Get("temp_ssh_key_id").(string)
	keys := append(c.SSHKeyIDs, tempKey)

	serverOpts := &govultr.ServerOptions{
		IsoID:                c.ISOID,
		SnapshotID:           c.SnapshotID,
		AppID:                c.AppID,
		ScriptID:             c.ScriptID,
		EnableIPV6:           c.EnableIPV6,
		EnablePrivateNetwork: c.EnablePrivateNetwork,
		Label:                c.Label,
		SSHKeyIDs:            keys,
		UserData:             c.UserData,
		NotifyActivate:       false,
		Hostname:             c.Hostname,
		Tag:                  c.Tag,
	}

	server, err := s.client.Server.Create(ctx, c.RegionID, c.PlanID, c.OSID, serverOpts)
	if err != nil {
		err = errors.New("Error creating server: " + err.Error())
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// wait until server is running
	ui.Say(fmt.Sprintf("Waiting %ds for server %s to power on...",
		int(c.stateTimeout/time.Second), server.InstanceID))

	err = waitForServerState("active", "running", server.InstanceID, s.client, c.stateTimeout)
	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	server, err = s.client.Server.GetServer(context.Background(), server.InstanceID)
	if err != nil {
		err := fmt.Errorf("Error getting server: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	state.Put("server", server)
	state.Put("server_ip", server.MainIP)
	state.Put("server_id", server.InstanceID)

	return multistep.ActionContinue
}

func (s *stepCreateServer) Cleanup(state multistep.StateBag) {
	server, ok := state.GetOk("server")
	if !ok {
		return
	}

	ui := state.Get("ui").(packer.Ui)
	instanceID := server.(*govultr.Server).InstanceID

	ui.Say("Destroying server " + instanceID)
	if err := s.client.Server.Delete(context.Background(), instanceID); err != nil {
		state.Put("error", err)
	}
}

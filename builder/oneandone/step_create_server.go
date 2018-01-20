package oneandone

import (
	"fmt"
	"strings"
	"time"

	"github.com/1and1/oneandone-cloudserver-sdk-go"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"strings"
	"time"
)

type stepCreateServer struct{}

func (s *stepCreateServer) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	c := state.Get("config").(*Config)

	if sshkey, ok := state.GetOk("publicKey"); ok {
		c.SSHKey = sshkey.(string)
	}

	token := oneandone.SetToken(c.Token)

	//Create an API client
	api := oneandone.New(token, c.Url)

	// List server appliances
	saps, _ := api.ListServerAppliances()

	time.Sleep(time.Second * 10)

	var sa oneandone.ServerAppliance
	for _, a := range saps {

		if a.Type == "IMAGE" && strings.Contains(strings.ToLower(a.Name), strings.ToLower(c.Image)) {
			sa = a
			break
		}
	}

	if c.DiskSize < sa.MinHddSize {
		ui.Error(fmt.Sprintf("Minimum required disk size %d", sa.MinHddSize))
	}

	ui.Say("Creating Server...")

	// Create a server
	req := oneandone.ServerRequest{
		Name:        c.SnapshotName,
		Description: "Example server description.",
		ApplianceId: sa.Id,
		PowerOn:     true,
		Hardware: oneandone.Hardware{
			Vcores:            1,
			CoresPerProcessor: 1,
			Ram:               2,
			Hdds: []oneandone.Hdd{
				{
					Size:   c.DiskSize,
					IsMain: true,
				},
			},
		},
	}

	if c.DataCenterId != "" {
		req.DatacenterId = c.DataCenterId
	}

	if c.Comm.SSHPassword != "" {
		req.Password = c.Comm.SSHPassword
	}
	if c.SSHKey != "" {
		req.SSHKey = c.SSHKey
	}

	server_id, server, err := api.CreateServer(&req)

	if err == nil {
		// Wait until server is created and powered on for at most 60 x 10 seconds
		err = api.WaitForState(server, "POWERED_ON", 10, c.Retries)
	} else {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Get a server
	server, err = api.GetServer(server_id)
	if err != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("server_id", server_id)

	state.Put("server_ip", server.Ips[0].Ip)

	return multistep.ActionContinue
}

func (s *stepCreateServer) Cleanup(state multistep.StateBag) {
	c := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Removing Server...")

	token := oneandone.SetToken(c.Token)
	//Create an API client
	api := oneandone.New(token, oneandone.BaseUrl)

	var serverId string
	if temp, ok := state.GetOk("server_id"); ok {
		serverId = temp.(string)
	}

	if serverId != "" {
		server, err := api.ShutdownServer(serverId, false)
		if err != nil {
			ui.Error(fmt.Sprintf("Error shutting down 1and1 server. Please destroy it manually: %s", serverId))
			ui.Error(err.Error())
		}
		err = api.WaitForState(server, "POWERED_OFF", 10, c.Retries)

		server, err = api.DeleteServer(server.Id, false)

		if err != nil {
			ui.Error(fmt.Sprintf("Error deleting 1and1 server. Please destroy it manually: %s", serverId))
			ui.Error(err.Error())
		}
	}
}

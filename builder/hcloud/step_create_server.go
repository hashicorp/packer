package hcloud

import (
	"context"
	"fmt"
	"io/ioutil"
	"sort"
	"strings"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

type stepCreateServer struct {
	serverId int
}

func (s *stepCreateServer) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("hcloudClient").(*hcloud.Client)
	ui := state.Get("ui").(packersdk.Ui)
	c := state.Get("config").(*Config)
	sshKeyId := state.Get("ssh_key_id").(int)

	// Create the server based on configuration
	ui.Say("Creating server...")

	userData := c.UserData
	if c.UserDataFile != "" {
		contents, err := ioutil.ReadFile(c.UserDataFile)
		if err != nil {
			state.Put("error", fmt.Errorf("Problem reading user data file: %s", err))
			return multistep.ActionHalt
		}

		userData = string(contents)
	}

	sshKeys := []*hcloud.SSHKey{{ID: sshKeyId}}
	for _, k := range c.SSHKeys {
		sshKey, _, err := client.SSHKey.Get(ctx, k)
		if err != nil {
			ui.Error(err.Error())
			state.Put("error", fmt.Errorf("Error fetching SSH key: %s", err))
			return multistep.ActionHalt
		}
		if sshKey == nil {
			state.Put("error", fmt.Errorf("Could not find key: %s", k))
			return multistep.ActionHalt
		}
		sshKeys = append(sshKeys, sshKey)
	}

	var image *hcloud.Image
	if c.Image != "" {
		image = &hcloud.Image{Name: c.Image}
	} else {
		var err error
		image, err = getImageWithSelectors(ctx, client, c)
		if err != nil {
			ui.Error(err.Error())
			state.Put("error", err)
			return multistep.ActionHalt
		}
		ui.Message(fmt.Sprintf("Using image %s with ID %d", image.Description, image.ID))
	}

	serverCreateResult, _, err := client.Server.Create(ctx, hcloud.ServerCreateOpts{
		Name:       c.ServerName,
		ServerType: &hcloud.ServerType{Name: c.ServerType},
		Image:      image,
		SSHKeys:    sshKeys,
		Location:   &hcloud.Location{Name: c.Location},
		UserData:   userData,
	})
	if err != nil {
		err := fmt.Errorf("Error creating server: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	state.Put("server_ip", serverCreateResult.Server.PublicNet.IPv4.IP.String())
	// We use this in cleanup
	s.serverId = serverCreateResult.Server.ID

	// Store the server id for later
	state.Put("server_id", serverCreateResult.Server.ID)
	// instance_id is the generic term used so that users can have access to the
	// instance id inside of the provisioners, used in step_provision.
	state.Put("instance_id", serverCreateResult.Server.ID)

	if err := waitForAction(ctx, client, serverCreateResult.Action); err != nil {
		err := fmt.Errorf("Error creating server: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	for _, nextAction := range serverCreateResult.NextActions {
		if err := waitForAction(ctx, client, nextAction); err != nil {
			err := fmt.Errorf("Error creating server: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	if c.RescueMode != "" {
		ui.Say("Enabling Rescue Mode...")
		rootPassword, err := setRescue(ctx, client, serverCreateResult.Server, c.RescueMode, sshKeys)
		if err != nil {
			err := fmt.Errorf("Error enabling rescue mode: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		ui.Say("Reboot server...")
		action, _, err := client.Server.Reset(ctx, serverCreateResult.Server)
		if err != nil {
			err := fmt.Errorf("Error rebooting server: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		if err := waitForAction(ctx, client, action); err != nil {
			err := fmt.Errorf("Error rebooting server: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		if c.RescueMode == "freebsd64" {
			// We will set this only on freebsd
			ui.Say("Using Root Password instead of SSH Keys...")
			c.Comm.SSHPassword = rootPassword
		}
	}

	return multistep.ActionContinue
}

func (s *stepCreateServer) Cleanup(state multistep.StateBag) {
	// If the serverID isn't there, we probably never created it
	if s.serverId == 0 {
		return
	}

	client := state.Get("hcloudClient").(*hcloud.Client)
	ui := state.Get("ui").(packersdk.Ui)

	// Destroy the server we just created
	ui.Say("Destroying server...")
	_, err := client.Server.Delete(context.TODO(), &hcloud.Server{ID: s.serverId})
	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error destroying server. Please destroy it manually: %s", err))
	}
}

func setRescue(ctx context.Context, client *hcloud.Client, server *hcloud.Server, rescue string, sshKeys []*hcloud.SSHKey) (string, error) {
	rescueChanged := false
	if server.RescueEnabled {
		rescueChanged = true
		action, _, err := client.Server.DisableRescue(ctx, server)
		if err != nil {
			return "", err
		}
		if err := waitForAction(ctx, client, action); err != nil {
			return "", err
		}
	}
	if rescue != "" {
		rescueChanged = true
		if rescue == "freebsd64" {
			sshKeys = nil // freebsd64 doesn't allow ssh keys so we will remove them here
		}
		res, _, err := client.Server.EnableRescue(ctx, server, hcloud.ServerEnableRescueOpts{
			Type:    hcloud.ServerRescueType(rescue),
			SSHKeys: sshKeys,
		})
		if err != nil {
			return "", err
		}
		if err := waitForAction(ctx, client, res.Action); err != nil {
			return "", err
		}
		return res.RootPassword, nil
	}
	if rescueChanged {
		action, _, err := client.Server.Reset(ctx, server)
		if err != nil {
			return "", err
		}
		if err := waitForAction(ctx, client, action); err != nil {
			return "", err
		}
	}
	return "", nil
}

func waitForAction(ctx context.Context, client *hcloud.Client, action *hcloud.Action) error {
	_, errCh := client.Action.WatchProgress(ctx, action)
	if err := <-errCh; err != nil {
		return err
	}
	return nil
}

func getImageWithSelectors(ctx context.Context, client *hcloud.Client, c *Config) (*hcloud.Image, error) {
	var allImages []*hcloud.Image

	var selector = strings.Join(c.ImageFilter.WithSelector, ",")
	opts := hcloud.ImageListOpts{
		ListOpts: hcloud.ListOpts{LabelSelector: selector},
		Status:   []hcloud.ImageStatus{hcloud.ImageStatusAvailable},
	}

	allImages, err := client.Image.AllWithOpts(ctx, opts)
	if err != nil {
		return nil, err
	}
	if len(allImages) == 0 {
		return nil, fmt.Errorf("no image found for selector %q", selector)
	}
	if len(allImages) > 1 {
		if !c.ImageFilter.MostRecent {
			return nil, fmt.Errorf("more than one image found for selector %q", selector)
		}

		sort.Slice(allImages, func(i, j int) bool {
			return allImages[i].Created.After(allImages[j].Created)
		})
	}

	return allImages[0], nil
}

package openstack

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/bootfromvolume"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/keypairs"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type StepRunSourceServer struct {
	Name                  string
	SecurityGroups        []string
	AvailabilityZone      string
	UserData              string
	UserDataFile          string
	ConfigDrive           bool
	InstanceMetadata      map[string]string
	UseBlockStorageVolume bool
	ForceDelete           bool
	server                *servers.Server
}

func (s *StepRunSourceServer) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	flavor := state.Get("flavor_id").(string)
	sourceImage := state.Get("source_image").(string)
	networks := state.Get("networks").([]servers.Network)
	ui := state.Get("ui").(packersdk.Ui)

	// We need the v2 compute client
	computeClient, err := config.computeV2Client()
	if err != nil {
		err = fmt.Errorf("Error initializing compute client: %s", err)
		state.Put("error", err)
		return multistep.ActionHalt
	}

	userData := []byte(s.UserData)
	if s.UserDataFile != "" {
		userData, err = ioutil.ReadFile(s.UserDataFile)
		if err != nil {
			err = fmt.Errorf("Error reading user data file: %s", err)
			state.Put("error", err)
			return multistep.ActionHalt
		}
	}

	ui.Say("Launching server...")

	serverOpts := servers.CreateOpts{
		Name:             s.Name,
		ImageRef:         sourceImage,
		FlavorRef:        flavor,
		SecurityGroups:   s.SecurityGroups,
		Networks:         networks,
		AvailabilityZone: s.AvailabilityZone,
		UserData:         userData,
		ConfigDrive:      &s.ConfigDrive,
		ServiceClient:    computeClient,
		Metadata:         s.InstanceMetadata,
	}

	var serverOptsExt servers.CreateOptsBuilder

	// Create root volume in the Block Storage service if required.
	// Add block device mapping v2 to the server create options if required.
	if s.UseBlockStorageVolume {
		volume := state.Get("volume_id").(string)
		blockDeviceMappingV2 := []bootfromvolume.BlockDevice{
			{
				BootIndex:       0,
				DestinationType: bootfromvolume.DestinationVolume,
				SourceType:      bootfromvolume.SourceVolume,
				UUID:            volume,
			},
		}
		// ImageRef and block device mapping is an invalid options combination.
		serverOpts.ImageRef = ""
		serverOptsExt = bootfromvolume.CreateOptsExt{
			CreateOptsBuilder: serverOpts,
			BlockDevice:       blockDeviceMappingV2,
		}
	} else {
		serverOptsExt = serverOpts
	}

	// Add keypair to the server create options.
	keyName := config.Comm.SSHKeyPairName
	if keyName != "" {
		serverOptsExt = keypairs.CreateOptsExt{
			CreateOptsBuilder: serverOptsExt,
			KeyName:           keyName,
		}
	}

	ui.Say("Launching server...")
	s.server, err = servers.Create(computeClient, serverOptsExt).Extract()
	if err != nil {
		err := fmt.Errorf("Error launching source server: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Message(fmt.Sprintf("Server ID: %s", s.server.ID))
	log.Printf("server id: %s", s.server.ID)

	ui.Say("Waiting for server to become ready...")
	stateChange := StateChangeConf{
		Pending:   []string{"BUILD"},
		Target:    []string{"ACTIVE"},
		Refresh:   ServerStateRefreshFunc(computeClient, s.server),
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
	// instance_id is the generic term used so that users can have access to the
	// instance id inside of the provisioners, used in step_provision.
	state.Put("instance_id", s.server.ID)

	return multistep.ActionContinue
}

func (s *StepRunSourceServer) Cleanup(state multistep.StateBag) {
	if s.server == nil {
		return
	}

	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packersdk.Ui)

	// We need the v2 compute client
	computeClient, err := config.computeV2Client()
	if err != nil {
		ui.Error(fmt.Sprintf("Error terminating server, may still be around: %s", err))
		return
	}

	ui.Say(fmt.Sprintf("Terminating the source server: %s ...", s.server.ID))
	if config.ForceDelete {
		if err := servers.ForceDelete(computeClient, s.server.ID).ExtractErr(); err != nil {
			ui.Error(fmt.Sprintf("Error terminating server, may still be around: %s", err))
			return
		}
	} else {
		if err := servers.Delete(computeClient, s.server.ID).ExtractErr(); err != nil {
			ui.Error(fmt.Sprintf("Error terminating server, may still be around: %s", err))
			return
		}
	}

	stateChange := StateChangeConf{
		Pending: []string{"ACTIVE", "BUILD", "REBUILD", "SUSPENDED", "SHUTOFF", "STOPPED"},
		Refresh: ServerStateRefreshFunc(computeClient, s.server),
		Target:  []string{"DELETED"},
	}

	WaitForState(&stateChange)
}

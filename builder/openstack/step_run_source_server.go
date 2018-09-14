package openstack

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/bootfromvolume"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/keypairs"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type StepRunSourceServer struct {
	Name                  string
	SourceImage           string
	SourceImageName       string
	SecurityGroups        []string
	Networks              []string
	Ports                 []string
	AvailabilityZone      string
	UserData              string
	UserDataFile          string
	ConfigDrive           bool
	InstanceMetadata      map[string]string
	UseBlockStorageVolume bool
	server                *servers.Server
	Comm                  *communicator.Config
}

func (s *StepRunSourceServer) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(Config)
	flavor := state.Get("flavor_id").(string)
	ui := state.Get("ui").(packer.Ui)

	// We need the v2 compute client
	computeClient, err := config.computeV2Client()
	if err != nil {
		err = fmt.Errorf("Error initializing compute client: %s", err)
		state.Put("error", err)
		return multistep.ActionHalt
	}

	networks := make([]servers.Network, len(s.Networks)+len(s.Ports))
	i := 0
	for ; i < len(s.Ports); i++ {
		networks[i].Port = s.Ports[i]
	}
	for ; i < len(networks); i++ {
		networks[i].UUID = s.Networks[i]
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
		ImageRef:         s.SourceImage,
		ImageName:        s.SourceImageName,
		FlavorRef:        flavor,
		SecurityGroups:   s.SecurityGroups,
		Networks:         networks,
		AvailabilityZone: s.AvailabilityZone,
		UserData:         userData,
		ConfigDrive:      &s.ConfigDrive,
		ServiceClient:    computeClient,
		Metadata:         s.InstanceMetadata,
	}

	// check if image filter returned a source image ID and replace
	if imageID, ok := state.GetOk("source_image"); ok {
		serverOpts.ImageRef = imageID.(string)
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
	keyName := s.Comm.SSHKeyPairName
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

	return multistep.ActionContinue
}

func (s *StepRunSourceServer) Cleanup(state multistep.StateBag) {
	if s.server == nil {
		return
	}

	config := state.Get("config").(Config)
	ui := state.Get("ui").(packer.Ui)

	// We need the v2 compute client
	computeClient, err := config.computeV2Client()
	if err != nil {
		ui.Error(fmt.Sprintf("Error terminating server, may still be around: %s", err))
		return
	}

	ui.Say(fmt.Sprintf("Terminating the source server: %s ...", s.server.ID))
	if err := servers.Delete(computeClient, s.server.ID).ExtractErr(); err != nil {
		ui.Error(fmt.Sprintf("Error terminating server, may still be around: %s", err))
		return
	}

	stateChange := StateChangeConf{
		Pending: []string{"ACTIVE", "BUILD", "REBUILD", "SUSPENDED", "SHUTOFF", "STOPPED"},
		Refresh: ServerStateRefreshFunc(computeClient, s.server),
		Target:  []string{"DELETED"},
	}

	WaitForState(&stateChange)
}

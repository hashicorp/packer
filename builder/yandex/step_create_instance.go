package yandex

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/c2h5oh/datasize"
	"github.com/hashicorp/packer/common/packerbuilderdata"
	"github.com/hashicorp/packer/common/uuid"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
)

const StandardImagesFolderID = "standard-images"

type StepCreateInstance struct {
	Debug         bool
	SerialLogFile string

	GeneratedData *packerbuilderdata.GeneratedData
}

func createNetwork(ctx context.Context, c *Config, d Driver) (*vpc.Network, error) {
	req := &vpc.CreateNetworkRequest{
		FolderId: c.FolderID,
		Name:     fmt.Sprintf("packer-network-%s", uuid.TimeOrderedUUID()),
	}

	sdk := d.SDK()

	op, err := sdk.WrapOperation(sdk.VPC().Network().Create(ctx, req))
	if err != nil {
		return nil, err
	}

	err = op.Wait(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := op.Response()
	if err != nil {
		return nil, err
	}

	network, ok := resp.(*vpc.Network)
	if !ok {
		return nil, errors.New("network create operation response doesn't contain Network")
	}
	return network, nil
}

func createSubnet(ctx context.Context, c *Config, d Driver, networkID string) (*vpc.Subnet, error) {
	req := &vpc.CreateSubnetRequest{
		FolderId:     c.FolderID,
		NetworkId:    networkID,
		Name:         fmt.Sprintf("packer-subnet-%s", uuid.TimeOrderedUUID()),
		ZoneId:       c.Zone,
		V4CidrBlocks: []string{"192.168.111.0/24"},
	}

	sdk := d.SDK()

	op, err := sdk.WrapOperation(sdk.VPC().Subnet().Create(ctx, req))
	if err != nil {
		return nil, err
	}

	err = op.Wait(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := op.Response()
	if err != nil {
		return nil, err
	}

	subnet, ok := resp.(*vpc.Subnet)
	if !ok {
		return nil, errors.New("subnet create operation response doesn't contain Subnet")
	}
	return subnet, nil
}

func getImage(ctx context.Context, c *Config, d Driver) (*Image, error) {
	if c.SourceImageID != "" {
		return d.GetImage(c.SourceImageID)
	}

	folderID := c.SourceImageFolderID
	if folderID == "" {
		folderID = StandardImagesFolderID
	}

	switch {
	case c.SourceImageFamily != "":
		return d.GetImageFromFolder(ctx, folderID, c.SourceImageFamily)
	case c.SourceImageName != "":
		return d.GetImageFromFolderByName(ctx, folderID, c.SourceImageName)
	}

	return &Image{}, errors.New("neither source_image_name nor source_image_family defined in config")
}

func (s *StepCreateInstance) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	sdk := state.Get("sdk").(*ycsdk.SDK)
	ui := state.Get("ui").(packer.Ui)
	config := state.Get("config").(*Config)
	driver := state.Get("driver").(Driver)

	ctx, cancel := context.WithTimeout(ctx, config.StateTimeout)
	defer cancel()

	sourceImage, err := getImage(ctx, config, driver)
	if err != nil {
		return stepHaltWithError(state, fmt.Errorf("Error getting source image for instance creation: %s", err))
	}

	if sourceImage.MinDiskSizeGb > config.DiskSizeGb {
		return stepHaltWithError(state, fmt.Errorf("Instance DiskSizeGb (%d) should be equal or greater "+
			"than SourceImage disk requirement (%d)", config.DiskSizeGb, sourceImage.MinDiskSizeGb))
	}

	ui.Say(fmt.Sprintf("Using as source image: %s (name: %q, family: %q)", sourceImage.ID, sourceImage.Name, sourceImage.Family))

	// create or reuse network configuration
	instanceSubnetID := ""
	if config.SubnetID == "" {
		// create Network and Subnet
		ui.Say("Creating network...")
		network, err := createNetwork(ctx, config, driver)
		if err != nil {
			return stepHaltWithError(state, fmt.Errorf("Error creating network: %s", err))
		}
		state.Put("network_id", network.Id)

		ui.Say(fmt.Sprintf("Creating subnet in zone %q...", config.Zone))
		subnet, err := createSubnet(ctx, config, driver, network.Id)
		if err != nil {
			return stepHaltWithError(state, fmt.Errorf("Error creating subnet: %s", err))
		}
		instanceSubnetID = subnet.Id
		// save for cleanup
		state.Put("subnet_id", subnet.Id)
	} else {
		ui.Say("Use provided subnet id " + config.SubnetID)
		instanceSubnetID = config.SubnetID
	}

	// Create an instance based on the configuration
	ui.Say("Creating instance...")

	instanceMetadata, err := config.createInstanceMetadata(string(config.Communicator.SSHPublicKey))
	if err != nil {
		return stepHaltWithError(state, fmt.Errorf("Error preparing instance metadata: %s", err))
	}

	// TODO make part metadata prepare process
	if config.UseIPv6 {
		// this ugly hack will replace user provided 'user-data'
		userData := `#cloud-config
runcmd:
- [ sh, -c, '/sbin/dhclient -6 -D LL -nw -pf /run/dhclient_ipv6.eth0.pid -lf /var/lib/dhcp/dhclient_ipv6.eth0.leases eth0' ]
`
		instanceMetadata["user-data"] = userData
	}

	req := &compute.CreateInstanceRequest{
		FolderId:   config.FolderID,
		Name:       config.InstanceName,
		Labels:     config.Labels,
		ZoneId:     config.Zone,
		PlatformId: config.PlatformID,
		SchedulingPolicy: &compute.SchedulingPolicy{
			Preemptible: config.Preemptible,
		},
		ResourcesSpec: &compute.ResourcesSpec{
			Memory: toBytes(config.InstanceMemory),
			Cores:  int64(config.InstanceCores),
			Gpus:   int64(config.InstanceGpus),
		},
		Metadata: instanceMetadata,
		BootDiskSpec: &compute.AttachedDiskSpec{
			AutoDelete: false,
			Disk: &compute.AttachedDiskSpec_DiskSpec_{
				DiskSpec: &compute.AttachedDiskSpec_DiskSpec{
					Name:   config.DiskName,
					TypeId: config.DiskType,
					Size:   int64((datasize.ByteSize(config.DiskSizeGb) * datasize.GB).Bytes()),
					Source: &compute.AttachedDiskSpec_DiskSpec_ImageId{
						ImageId: sourceImage.ID,
					},
				},
			},
		},
		NetworkInterfaceSpecs: []*compute.NetworkInterfaceSpec{
			{
				SubnetId:             instanceSubnetID,
				PrimaryV4AddressSpec: &compute.PrimaryAddressSpec{},
			},
		},
	}

	if config.ServiceAccountID != "" {
		req.ServiceAccountId = config.ServiceAccountID
	}

	if config.UseIPv6 {
		req.NetworkInterfaceSpecs[0].PrimaryV6AddressSpec = &compute.PrimaryAddressSpec{}
	}

	if config.UseIPv4Nat {
		req.NetworkInterfaceSpecs[0].PrimaryV4AddressSpec = &compute.PrimaryAddressSpec{
			OneToOneNatSpec: &compute.OneToOneNatSpec{
				IpVersion: compute.IpVersion_IPV4,
			},
		}
	}

	op, err := sdk.WrapOperation(sdk.Compute().Instance().Create(ctx, req))
	if err != nil {
		return stepHaltWithError(state, fmt.Errorf("Error create instance: %s", err))
	}

	opMetadata, err := op.Metadata()
	if err != nil {
		return stepHaltWithError(state, fmt.Errorf("Error get create operation metadata: %s", err))
	}

	if cimd, ok := opMetadata.(*compute.CreateInstanceMetadata); ok {
		state.Put("instance_id", cimd.InstanceId)
	} else {
		return stepHaltWithError(state, fmt.Errorf("could not get Instance ID from operation metadata"))
	}

	err = op.Wait(ctx)
	if err != nil {
		return stepHaltWithError(state, fmt.Errorf("Error create instance: %s", err))
	}

	resp, err := op.Response()
	if err != nil {
		return stepHaltWithError(state, err)
	}

	instance, ok := resp.(*compute.Instance)
	if !ok {
		return stepHaltWithError(state, fmt.Errorf("response doesn't contain Instance"))
	}

	state.Put("disk_id", instance.BootDisk.DiskId)
	// instance_id is the generic term used so that users can have access to the
	// instance id inside of the provisioners, used in step_provision.
	state.Put("instance_id", instance.Id)

	if s.Debug {
		ui.Message(fmt.Sprintf("Instance ID %s started. Current instance status %s", instance.Id, instance.Status))
		ui.Message(fmt.Sprintf("Disk ID %s. ", instance.BootDisk.DiskId))
	}

	// provision generated_data from declared in Builder.Prepare func
	// see doc https://www.packer.io/docs/extending/custom-builders#build-variables for details
	s.GeneratedData.Put("SourceImageID", sourceImage.ID)
	s.GeneratedData.Put("SourceImageName", sourceImage.Name)
	s.GeneratedData.Put("SourceImageDescription", sourceImage.Description)
	s.GeneratedData.Put("SourceImageFamily", sourceImage.Family)
	s.GeneratedData.Put("SourceImageFolderID", sourceImage.FolderID)

	return multistep.ActionContinue
}

func (s *StepCreateInstance) Cleanup(state multistep.StateBag) {
	config := state.Get("config").(*Config)
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	ctx, cancel := context.WithTimeout(context.Background(), config.StateTimeout)
	defer cancel()

	if s.SerialLogFile != "" {
		ui.Say("Current state 'cancelled' or 'halted'...")
		err := s.writeSerialLogFile(ctx, state)
		if err != nil {
			ui.Error(err.Error())
		}
	}

	instanceIDRaw, ok := state.GetOk("instance_id")
	if ok {
		instanceID := instanceIDRaw.(string)
		if instanceID != "" {
			ui.Say("Destroying instance...")
			err := driver.DeleteInstance(ctx, instanceID)
			if err != nil {
				ui.Error(fmt.Sprintf(
					"Error destroying instance (id: %s). Please destroy it manually: %s", instanceID, err))
			}
			ui.Message("Instance has been destroyed!")
		}
	}

	subnetIDRaw, ok := state.GetOk("subnet_id")
	if ok {
		subnetID := subnetIDRaw.(string)
		if subnetID != "" {
			// Destroy the subnet we just created
			ui.Say("Destroying subnet...")
			err := driver.DeleteSubnet(ctx, subnetID)
			if err != nil {
				ui.Error(fmt.Sprintf(
					"Error destroying subnet (id: %s). Please destroy it manually: %s", subnetID, err))
			}
			ui.Message("Subnet has been deleted!")
		}
	}

	// Destroy the network we just created
	networkIDRaw, ok := state.GetOk("network_id")
	if ok {
		networkID := networkIDRaw.(string)
		if networkID != "" {
			// Destroy the network we just created
			ui.Say("Destroying network...")
			err := driver.DeleteNetwork(ctx, networkID)
			if err != nil {
				ui.Error(fmt.Sprintf(
					"Error destroying network (id: %s). Please destroy it manually: %s", networkID, err))
			}
			ui.Message("Network has been deleted!")
		}
	}

	diskIDRaw, ok := state.GetOk("disk_id")
	if ok {
		ui.Say("Destroying boot disk...")
		diskID := diskIDRaw.(string)
		err := driver.DeleteDisk(ctx, diskID)
		if err != nil {
			ui.Error(fmt.Sprintf(
				"Error destroying boot disk (id: %s). Please destroy it manually: %s", diskID, err))
		}
		ui.Message("Disk has been deleted!")
	}
}

func (s *StepCreateInstance) writeSerialLogFile(ctx context.Context, state multistep.StateBag) error {
	sdk := state.Get("sdk").(*ycsdk.SDK)
	ui := state.Get("ui").(packer.Ui)

	instanceID := state.Get("instance_id").(string)
	ui.Say("Try get instance's serial port output and write to file " + s.SerialLogFile)
	serialOutput, err := sdk.Compute().Instance().GetSerialPortOutput(ctx, &compute.GetInstanceSerialPortOutputRequest{
		InstanceId: instanceID,
	})
	if err != nil {
		return fmt.Errorf("Failed to get serial port output for instance (id: %s): %s", instanceID, err)
	}
	if err := ioutil.WriteFile(s.SerialLogFile, []byte(serialOutput.Contents), 0600); err != nil {
		return fmt.Errorf("Failed to write serial port output to file: %s", err)
	}
	ui.Message("Serial port output has been successfully written")
	return nil
}

func (c *Config) createInstanceMetadata(sshPublicKey string) (map[string]string, error) {
	instanceMetadata := make(map[string]string)

	// Copy metadata from config.
	for k, file := range c.MetadataFromFile {
		contents, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("error while read file '%s' with content for value of metadata key '%s': %s", file, k, err)
		}
		instanceMetadata[k] = string(contents)
	}

	for k, v := range c.Metadata {
		instanceMetadata[k] = v
	}

	if sshPublicKey != "" {
		sshMetaKey := "ssh-keys"
		sshKeys := fmt.Sprintf("%s:%s", c.Communicator.SSHUsername, sshPublicKey)
		if confSSHKeys, exists := instanceMetadata[sshMetaKey]; exists {
			sshKeys = fmt.Sprintf("%s\n%s", sshKeys, confSSHKeys)
		}
		instanceMetadata[sshMetaKey] = sshKeys
	}

	return instanceMetadata, nil
}

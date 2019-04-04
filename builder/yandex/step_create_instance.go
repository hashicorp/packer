package yandex

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/c2h5oh/datasize"
	"github.com/hashicorp/packer/common/uuid"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
)

type stepCreateInstance struct {
	Debug             bool
	SerialLogFile     string
	cleanupInstanceID string
	cleanupNetworkID  string
	cleanupSubnetID   string
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

	network, ok := resp.(*vpc.Subnet)
	if !ok {
		return nil, errors.New("subnet create operation response doesn't contain Network")
	}
	return network, nil
}

func getImage(ctx context.Context, c *Config, d Driver) (*Image, error) {
	if c.SourceImageID != "" {
		return d.GetImage(c.SourceImageID)
	}

	familyName := c.SourceImageFamily
	if c.SourceImageFolderID != "" {
		return d.GetImageFromFolder(ctx, c.SourceImageFolderID, familyName)
	}
	return d.GetImageFromFolder(ctx, "standard-images", familyName)
}

func (s *stepCreateInstance) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	sdk := state.Get("sdk").(*ycsdk.SDK)
	ui := state.Get("ui").(packer.Ui)
	c := state.Get("config").(*Config)
	d := state.Get("driver").(Driver)

	ctx, cancel := context.WithTimeout(ctx, c.StateTimeout)
	defer cancel()

	// create or reuse network configuration
	instanceSubnetID := ""
	if c.SubnetID == "" {
		// create Network and Subnet
		ui.Say("Creating network...")
		network, err := createNetwork(ctx, c, d)
		if err != nil {
			return stepHaltWithError(state, fmt.Errorf("Error creating network: %s", err))
		}
		state.Put("network_id", network.Id)
		s.cleanupNetworkID = network.Id

		ui.Say(fmt.Sprintf("Creating subnet in zone %q...", c.Zone))
		subnet, err := createSubnet(ctx, c, d, network.Id)
		if err != nil {
			return stepHaltWithError(state, fmt.Errorf("Error creating subnet: %s", err))
		}
		state.Put("subnet_id", subnet.Id)
		instanceSubnetID = subnet.Id
		// save for cleanup
		s.cleanupSubnetID = subnet.Id
	} else {
		ui.Say("Use provided subnet id " + c.SubnetID)
		instanceSubnetID = c.SubnetID
	}

	// Create an instance based on the configuration
	ui.Say("Creating instance...")
	sourceImage, err := getImage(ctx, c, d)
	if err != nil {
		return stepHaltWithError(state, fmt.Errorf("Error getting source image for instance creation: %s", err))
	}

	if sourceImage.MinDiskSizeGb > c.DiskSizeGb {
		return stepHaltWithError(state, fmt.Errorf("Instance DiskSizeGb (%d) should be equal or greater "+
			"than SourceImage disk requirement (%d)", c.DiskSizeGb, sourceImage.MinDiskSizeGb))
	}

	instanceMetadata, err := c.createInstanceMetadata(string(c.Communicator.SSHPublicKey))
	if err != nil {
		return stepHaltWithError(state, fmt.Errorf("instance metadata prepare error: %s", err))
	}

	// TODO make part metadata prepare process
	if c.UseIPv6 {
		// this ugly hack will replace user provided 'user-data'
		userData := `#cloud-config
runcmd:
- [ sh, -c, '/sbin/dhclient -6 -D LL -nw -pf /run/dhclient_ipv6.eth0.pid -lf /var/lib/dhcp/dhclient_ipv6.eth0.leases eth0' ]
`
		instanceMetadata["user-data"] = userData
	}

	req := &compute.CreateInstanceRequest{
		FolderId:   c.FolderID,
		Name:       c.InstanceName,
		Labels:     c.Labels,
		ZoneId:     c.Zone,
		PlatformId: "standard-v1",
		ResourcesSpec: &compute.ResourcesSpec{
			Memory: toBytes(c.InstanceMemory),
			Cores:  int64(c.InstanceCores),
		},
		Metadata: instanceMetadata,
		BootDiskSpec: &compute.AttachedDiskSpec{
			AutoDelete: false,
			Disk: &compute.AttachedDiskSpec_DiskSpec_{
				DiskSpec: &compute.AttachedDiskSpec_DiskSpec{
					Name:   c.DiskName,
					TypeId: c.DiskType,
					Size:   int64((datasize.ByteSize(c.DiskSizeGb) * datasize.GB).Bytes()),
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

	if c.UseIPv6 {
		req.NetworkInterfaceSpecs[0].PrimaryV6AddressSpec = &compute.PrimaryAddressSpec{}
	}

	if c.UseIPv4Nat {
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

	// We use this in cleanup
	s.cleanupInstanceID = instance.Id

	if s.Debug {
		ui.Message(fmt.Sprintf("Instance ID %s started. Current instance status %s", instance.Id, instance.Status))
	}

	// Store the instance id for later
	state.Put("instance_id", instance.Id)
	state.Put("disk_id", instance.BootDisk.DiskId)

	return multistep.ActionContinue
}

func (s *stepCreateInstance) Cleanup(state multistep.StateBag) {
	// If the cleanupInstanceID isn't there, we probably never created it
	if s.cleanupInstanceID == "" {
		return
	}

	c := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)

	if s.SerialLogFile != "" {
		ui.Say("Current state 'cancelled' or 'halted'...")
		err := s.writeSerialLogFile(state)
		if err != nil {
			ui.Error(err.Error())
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.StateTimeout)
	defer cancel()

	if s.cleanupSubnetID != "" {
		// Destroy the subnet we just created
		ui.Say("Destroying subnet...")
		err := deleteSubnet(ctx, s.cleanupSubnetID, state)
		if err != nil {
			ui.Error(fmt.Sprintf(
				"Error destroying subnet (id: %s). Please destroy it manually: %s", s.cleanupSubnetID, err))
		}

		// some sleep before delete network
		time.Sleep(10 * time.Second)

		// Destroy the network we just created
		ui.Say("Destroying network...")
		err = deleteNetwork(ctx, s.cleanupNetworkID, state)
		if err != nil {
			ui.Error(fmt.Sprintf(
				"Error destroying network (id: %s). Please destroy it manually: %s", s.cleanupNetworkID, err))
		}
	}

	ui.Say("Destroying boot disk...")
	diskID := state.Get("disk_id").(string)
	err := deleteDisk(ctx, diskID, state)
	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error destroying boot disk (id: %s). Please destroy it manually: %s", s.cleanupNetworkID, err))
	}
}

func deleteSubnet(ctx context.Context, subnetID string, state multistep.StateBag) error {
	sdk := state.Get("sdk").(*ycsdk.SDK)

	op, err := sdk.WrapOperation(sdk.VPC().Subnet().Delete(ctx, &vpc.DeleteSubnetRequest{
		SubnetId: subnetID,
	}))
	if err != nil {
		return err
	}

	err = op.Wait(ctx)
	if err != nil {
		return err
	}

	_, err = op.Response()
	return err
}

func deleteNetwork(ctx context.Context, networkID string, state multistep.StateBag) error {
	sdk := state.Get("sdk").(*ycsdk.SDK)

	op, err := sdk.WrapOperation(sdk.VPC().Network().Delete(ctx, &vpc.DeleteNetworkRequest{
		NetworkId: networkID,
	}))
	if err != nil {
		return err
	}

	err = op.Wait(ctx)
	if err != nil {
		return err
	}

	_, err = op.Response()
	return err
}

func deleteDisk(ctx context.Context, diskID string, state multistep.StateBag) error {
	sdk := state.Get("sdk").(*ycsdk.SDK)

	op, err := sdk.WrapOperation(sdk.Compute().Disk().Delete(ctx, &compute.DeleteDiskRequest{
		DiskId: diskID,
	}))
	if err != nil {
		return err
	}

	err = op.Wait(ctx)
	if err != nil {
		return err
	}

	_, err = op.Response()
	return err
}

func (s *stepCreateInstance) writeSerialLogFile(state multistep.StateBag) error {
	sdk := state.Get("sdk").(*ycsdk.SDK)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Try get serial port output to file " + s.SerialLogFile)
	serialOutput, err := sdk.Compute().Instance().GetSerialPortOutput(context.Background(), &compute.GetInstanceSerialPortOutputRequest{
		InstanceId: s.cleanupInstanceID,
	})
	if err != nil {
		return fmt.Errorf("Failed to get serial port output for instance (id: %s): %s", s.cleanupInstanceID, err)
	}
	if err := ioutil.WriteFile(s.SerialLogFile, []byte(serialOutput.Contents), 0600); err != nil {
		return fmt.Errorf("Failed to write serial port output to file: %s", err)
	}
	ui.Message("Serial port output has been successfully written")
	return nil
}

func (c *Config) createInstanceMetadata(sshPublicKey string) (map[string]string, error) {
	instanceMetadata := make(map[string]string)
	var err error

	// Copy metadata from config.
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

	return instanceMetadata, err
}

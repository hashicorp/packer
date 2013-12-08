// Copyright (c) 2013 Kelsey Hightower. All rights reserved.
// Use of this source code is governed by the Apache License, Version 2.0
// that can be found in the LICENSE file.

package googlecompute

import (
	"fmt"

	"code.google.com/p/google-api-go-client/compute/v1beta16"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/common/uuid"
	"github.com/mitchellh/packer/packer"
)

// stepCreateInstance represents a Packer build step that creates GCE instances.
type stepCreateInstance struct {
	instanceName string
}

// Run executes the Packer build step that creates a GCE instance.
func (s *stepCreateInstance) Run(state multistep.StateBag) multistep.StepAction {
	var (
		client = state.Get("client").(*GoogleComputeClient)
		config = state.Get("config").(config)
		ui     = state.Get("ui").(packer.Ui)
	)
	ui.Say("Creating instance...")
	name := fmt.Sprintf("packer-%s", uuid.TimeOrderedUUID())
	// Build up the instance config.
	instanceConfig := &InstanceConfig{
		Description: "New instance created by Packer",
		Name:        name,
	}
	// Validate the zone.
	zone, err := client.GetZone(config.Zone)
	if err != nil {
		err := fmt.Errorf("Error creating instance: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	// Set the source image. Must be a fully-qualified URL.
	image, err := client.GetImage(config.SourceImage)
	if err != nil {
		err := fmt.Errorf("Error creating instance: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	instanceConfig.Image = image.SelfLink
	// Set the machineType. Must be a fully-qualified URL.
	machineType, err := client.GetMachineType(config.MachineType, zone.Name)
	if err != nil {
		err := fmt.Errorf("Error creating instance: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt

	}
	instanceConfig.MachineType = machineType.SelfLink
	// Set up the Network Interface.
	network, err := client.GetNetwork(config.Network)
	if err != nil {
		err := fmt.Errorf("Error creating instance: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	networkInterface := NewNetworkInterface(network, true)
	networkInterfaces := []*compute.NetworkInterface{
		networkInterface,
	}
	instanceConfig.NetworkInterfaces = networkInterfaces
	// Add the metadata, which also setups up the ssh key.
	metadata := make(map[string]string)
	sshPublicKey := state.Get("ssh_public_key").(string)
	metadata["sshKeys"] = fmt.Sprintf("%s:%s", config.SSHUsername, sshPublicKey)
	instanceConfig.Metadata = MapToMetadata(metadata)
	// Add the default service so we can create an image of the machine and
	// upload it to cloud storage.
	defaultServiceAccount := NewServiceAccount("default")
	serviceAccounts := []*compute.ServiceAccount{
		defaultServiceAccount,
	}
	instanceConfig.ServiceAccounts = serviceAccounts
	// Create the instance based on configuration
	operation, err := client.CreateInstance(zone.Name, instanceConfig)
	if err != nil {
		err := fmt.Errorf("Error creating instance: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	ui.Say("Waiting for the instance to be created...")
	err = waitForZoneOperationState("DONE", config.Zone, operation.Name, client, config.stateTimeout)
	if err != nil {
		err := fmt.Errorf("Error creating instance: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	// Update the state.
	state.Put("instance_name", name)
	s.instanceName = name
	return multistep.ActionContinue
}

// Cleanup destroys the GCE instance created during the image creation process.
func (s *stepCreateInstance) Cleanup(state multistep.StateBag) {
	var (
		client = state.Get("client").(*GoogleComputeClient)
		config = state.Get("config").(config)
		ui     = state.Get("ui").(packer.Ui)
	)
	if s.instanceName == "" {
		return
	}
	ui.Say("Destroying instance...")
	operation, err := client.DeleteInstance(config.Zone, s.instanceName)
	if err != nil {
		ui.Error(fmt.Sprintf("Error destroying instance. Please destroy it manually: %v", s.instanceName))
	}
	ui.Say("Waiting for the instance to be deleted...")
	for {
		status, err := client.ZoneOperationStatus(config.Zone, operation.Name)
		if err != nil {
			ui.Error(fmt.Sprintf("Error destroying instance. Please destroy it manually: %v", s.instanceName))
		}
		if status == "DONE" {
			break
		}
	}
	return
}

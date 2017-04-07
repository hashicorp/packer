// Copyright (c) 2017 Oracle America, Inc.
// The contents of this file are subject to the Mozilla Public License Version
// 2.0 (the "License"); you may not use this file except in compliance with the
// License. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/

package bmcs

import (
	"errors"
	"fmt"

	client "github.com/hashicorp/packer/builder/oracle/bmcs/client"
)

// driverBMCS implements the Driver interface and communicates with Oracle
// BMCS.
type driverBMCS struct {
	client *client.Client
	cfg    *Config
}

// NewDriverBMCS Creates a new driverBMCS with a connected client.
func NewDriverBMCS(cfg *Config) (Driver, error) {
	client, err := client.NewClient(cfg.AccessCfg)
	if err != nil {
		return nil, err
	}
	return &driverBMCS{client: client, cfg: cfg}, nil
}

// CreateInstance creates a new compute instance.
func (d *driverBMCS) CreateInstance(publicKey string) (string, error) {
	params := &client.LaunchInstanceParams{
		AvailabilityDomain: d.cfg.AvailabilityDomain,
		CompartmentID:      d.cfg.CompartmentID,
		ImageID:            d.cfg.BaseImageID,
		Shape:              d.cfg.Shape,
		SubnetID:           d.cfg.SubnetID,
		Metadata: map[string]string{
			"ssh_authorized_keys": publicKey,
		},
	}
	instance, err := d.client.Compute.Instances.Launch(params)
	if err != nil {
		return "", err
	}

	return instance.ID, nil
}

// CreateImage creates a new custom image.
func (d *driverBMCS) CreateImage(id string) (client.Image, error) {
	params := &client.CreateImageParams{
		CompartmentID: d.cfg.CompartmentID,
		InstanceID:    id,
		DisplayName:   d.cfg.ImageName,
	}
	image, err := d.client.Compute.Images.Create(params)
	if err != nil {
		return client.Image{}, err
	}

	return image, nil
}

// DeleteImage deletes a custom image.
func (d *driverBMCS) DeleteImage(id string) error {
	return d.client.Compute.Images.Delete(&client.DeleteImageParams{ID: id})
}

// GetInstanceIP returns the public IP corresponding to the given instance id.
func (d *driverBMCS) GetInstanceIP(id string) (string, error) {
	// get nvic and cross ref to find pub ip address
	vnics, err := d.client.Compute.VNICAttachments.List(
		&client.ListVnicAttachmentsParams{
			InstanceID:    id,
			CompartmentID: d.cfg.CompartmentID,
		},
	)
	if err != nil {
		return "", err
	}

	if len(vnics) < 1 {
		return "", errors.New("instance has zero VNICs")
	}

	vnic, err := d.client.Compute.VNICs.Get(&client.GetVNICParams{ID: vnics[0].VNICID})
	if err != nil {
		return "", fmt.Errorf("Error getting VNIC details: %s", err)
	}

	return vnic.PublicIP, nil
}

// TerminateInstance terminates a compute instance.
func (d *driverBMCS) TerminateInstance(id string) error {
	params := &client.TerminateInstanceParams{ID: id}
	return d.client.Compute.Instances.Terminate(params)
}

// WaitForImageCreation waits for a provisioning custom image to reach the
// "AVAILABLE" state.
func (d *driverBMCS) WaitForImageCreation(id string) error {
	return client.NewWaiter().WaitForResourceToReachState(
		d.client.Compute.Images,
		id,
		[]string{"PROVISIONING"},
		"AVAILABLE",
	)
}

// WaitForInstanceState waits for an instance to reach the a given terminal
// state.
func (d *driverBMCS) WaitForInstanceState(id string, waitStates []string, terminalState string) error {
	return client.NewWaiter().WaitForResourceToReachState(
		d.client.Compute.Instances,
		id,
		waitStates,
		terminalState,
	)
}

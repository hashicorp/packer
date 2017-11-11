package oci

import (
	"errors"
	"fmt"

	client "github.com/hashicorp/packer/builder/oracle/oci/client"
)

// driverOCI implements the Driver interface and communicates with Oracle
// OCI.
type driverOCI struct {
	client *client.Client
	cfg    *Config
}

// NewDriverOCI Creates a new driverOCI with a connected client.
func NewDriverOCI(cfg *Config) (Driver, error) {
	client, err := client.NewClient(cfg.AccessCfg)
	if err != nil {
		return nil, err
	}
	return &driverOCI{client: client, cfg: cfg}, nil
}

// CreateInstance creates a new compute instance.
func (d *driverOCI) CreateInstance(publicKey string) (string, error) {
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
func (d *driverOCI) CreateImage(id string) (client.Image, error) {
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
func (d *driverOCI) DeleteImage(id string) error {
	return d.client.Compute.Images.Delete(&client.DeleteImageParams{ID: id})
}

// GetInstanceIP returns the public IP corresponding to the given instance id.
func (d *driverOCI) GetInstanceIP(id string) (string, error) {
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
func (d *driverOCI) TerminateInstance(id string) error {
	params := &client.TerminateInstanceParams{ID: id}
	return d.client.Compute.Instances.Terminate(params)
}

// WaitForImageCreation waits for a provisioning custom image to reach the
// "AVAILABLE" state.
func (d *driverOCI) WaitForImageCreation(id string) error {
	return client.NewWaiter().WaitForResourceToReachState(
		d.client.Compute.Images,
		id,
		[]string{"PROVISIONING"},
		"AVAILABLE",
	)
}

// WaitForInstanceState waits for an instance to reach the a given terminal
// state.
func (d *driverOCI) WaitForInstanceState(id string, waitStates []string, terminalState string) error {
	return client.NewWaiter().WaitForResourceToReachState(
		d.client.Compute.Instances,
		id,
		waitStates,
		terminalState,
	)
}

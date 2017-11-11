package oci

import (
	client "github.com/hashicorp/packer/builder/oracle/oci/client"
)

// driverMock implements the Driver interface and communicates with Oracle
// OCI.
type driverMock struct {
	CreateInstanceID  string
	CreateInstanceErr error

	CreateImageID  string
	CreateImageErr error

	DeleteImageID  string
	DeleteImageErr error

	GetInstanceIPErr error

	TerminateInstanceID  string
	TerminateInstanceErr error

	WaitForImageCreationErr error

	WaitForInstanceStateErr error
}

// CreateInstance creates a new compute instance.
func (d *driverMock) CreateInstance(publicKey string) (string, error) {
	if d.CreateInstanceErr != nil {
		return "", d.CreateInstanceErr
	}

	d.CreateInstanceID = "ocid1..."

	return d.CreateInstanceID, nil
}

// CreateImage creates a new custom image.
func (d *driverMock) CreateImage(id string) (client.Image, error) {
	if d.CreateImageErr != nil {
		return client.Image{}, d.CreateImageErr
	}
	d.CreateImageID = id
	return client.Image{ID: id}, nil
}

// DeleteImage mocks deleting a custom image.
func (d *driverMock) DeleteImage(id string) error {
	if d.DeleteImageErr != nil {
		return d.DeleteImageErr
	}

	d.DeleteImageID = id

	return nil
}

// GetInstanceIP returns the public IP corresponding to the given instance id.
func (d *driverMock) GetInstanceIP(id string) (string, error) {
	if d.GetInstanceIPErr != nil {
		return "", d.GetInstanceIPErr
	}
	return "ip", nil
}

// TerminateInstance terminates a compute instance.
func (d *driverMock) TerminateInstance(id string) error {
	if d.TerminateInstanceErr != nil {
		return d.TerminateInstanceErr
	}

	d.TerminateInstanceID = id

	return nil
}

// WaitForImageCreation waits for a provisioning custom image to reach the
// "AVAILABLE" state.
func (d *driverMock) WaitForImageCreation(id string) error {
	return d.WaitForImageCreationErr
}

// WaitForInstanceState waits for an instance to reach the a given terminal
// state.
func (d *driverMock) WaitForInstanceState(id string, waitStates []string, terminalState string) error {
	return d.WaitForInstanceStateErr
}

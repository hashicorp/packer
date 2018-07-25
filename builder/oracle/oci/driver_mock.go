package oci

import (
	"context"

	"github.com/oracle/oci-go-sdk/core"
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

	cfg *Config
}

// CreateInstance creates a new compute instance.
func (d *driverMock) CreateInstance(ctx context.Context, publicKey string) (string, error) {
	if d.CreateInstanceErr != nil {
		return "", d.CreateInstanceErr
	}

	d.CreateInstanceID = "ocid1..."

	return d.CreateInstanceID, nil
}

// CreateImage creates a new custom image.
func (d *driverMock) CreateImage(ctx context.Context, id string) (core.Image, error) {
	if d.CreateImageErr != nil {
		return core.Image{}, d.CreateImageErr
	}
	d.CreateImageID = id
	return core.Image{Id: &id}, nil
}

// DeleteImage mocks deleting a custom image.
func (d *driverMock) DeleteImage(ctx context.Context, id string) error {
	if d.DeleteImageErr != nil {
		return d.DeleteImageErr
	}

	d.DeleteImageID = id

	return nil
}

// GetInstanceIP returns the public or private IP corresponding to the given instance id.
func (d *driverMock) GetInstanceIP(ctx context.Context, id string) (string, error) {
	if d.GetInstanceIPErr != nil {
		return "", d.GetInstanceIPErr
	}
	if d.cfg.UsePrivateIP {
		return "private_ip", nil
	}
	return "ip", nil
}

// TerminateInstance terminates a compute instance.
func (d *driverMock) TerminateInstance(ctx context.Context, id string) error {
	if d.TerminateInstanceErr != nil {
		return d.TerminateInstanceErr
	}

	d.TerminateInstanceID = id

	return nil
}

// WaitForImageCreation waits for a provisioning custom image to reach the
// "AVAILABLE" state.
func (d *driverMock) WaitForImageCreation(ctx context.Context, id string) error {
	return d.WaitForImageCreationErr
}

// WaitForInstanceState waits for an instance to reach the a given terminal
// state.
func (d *driverMock) WaitForInstanceState(ctx context.Context, id string, waitStates []string, terminalState string) error {
	return d.WaitForInstanceStateErr
}

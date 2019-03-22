package packer

import (
	"context"
	"time"
)

// MockProvisioner is an implementation of Provisioner that can be
// used for tests.
type MockProvisioner struct {
	ProvFunc func() error

	PrepCalled       bool
	PrepConfigs      []interface{}
	ProvCalled       bool
	ProvCommunicator Communicator
	ProvUi           Ui
	CancelCalled     bool
}

func (t *MockProvisioner) Prepare(configs ...interface{}) error {
	t.PrepCalled = true
	t.PrepConfigs = configs
	return nil
}

func (t *MockProvisioner) Provision(ctx context.Context, ui Ui, comm Communicator) error {
	go func() {
		select {
		case <-time.After(2 * time.Minute):
		case <-ctx.Done():
			t.CancelCalled = true
		}
	}()

	t.ProvCalled = true
	t.ProvCommunicator = comm
	t.ProvUi = ui

	if t.ProvFunc == nil {
		return nil
	}

	return t.ProvFunc()
}

func (t *MockProvisioner) Communicator() Communicator {
	return t.ProvCommunicator
}

func (t *MockProvisioner) ElevatedUser() string {
	return "user"
}

func (t *MockProvisioner) ElevatedPassword() string {
	return "password"
}

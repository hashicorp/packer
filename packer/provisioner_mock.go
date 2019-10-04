package packer

import (
	"context"
)

// MockProvisioner is an implementation of Provisioner that can be
// used for tests.
type MockProvisioner struct {
	ProvFunc func(context.Context) error

	PrepCalled       bool
	PrepConfigs      []interface{}
	ProvCalled       bool
	ProvCommunicator Communicator
	ProvUi           Ui
}

func (t *MockProvisioner) Prepare(configs ...interface{}) error {
	t.PrepCalled = true
	t.PrepConfigs = configs
	return nil
}

func (t *MockProvisioner) Provision(ctx context.Context, ui Ui, comm Communicator, generatedData interface{}) error {
	t.ProvCalled = true
	t.ProvCommunicator = comm
	t.ProvUi = ui

	if t.ProvFunc == nil {
		return nil
	}

	return t.ProvFunc(ctx)
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

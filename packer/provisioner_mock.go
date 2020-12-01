package packer

import (
	"context"

	"github.com/hashicorp/hcl/v2/hcldec"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

// MockProvisioner is an implementation of Provisioner that can be
// used for tests.
type MockProvisioner struct {
	ProvFunc func(context.Context) error

	PrepCalled       bool
	PrepConfigs      []interface{}
	ProvCalled       bool
	ProvRetried      bool
	ProvCommunicator packersdk.Communicator
	ProvUi           packersdk.Ui
}

func (tp *MockProvisioner) ConfigSpec() hcldec.ObjectSpec { return tp.FlatMapstructure().HCL2Spec() }

func (tp *MockProvisioner) FlatConfig() interface{} { return tp.FlatMapstructure() }

func (t *MockProvisioner) Prepare(configs ...interface{}) error {
	t.PrepCalled = true
	t.PrepConfigs = configs
	return nil
}

func (t *MockProvisioner) Provision(ctx context.Context, ui packersdk.Ui, comm packersdk.Communicator, generatedData map[string]interface{}) error {
	if t.ProvCalled {
		t.ProvRetried = true
		return nil
	}

	t.ProvCalled = true
	t.ProvCommunicator = comm
	t.ProvUi = ui

	if t.ProvFunc == nil {
		return nil
	}

	return t.ProvFunc(ctx)
}

func (t *MockProvisioner) Communicator() packersdk.Communicator {
	return t.ProvCommunicator
}

func (t *MockProvisioner) ElevatedUser() string {
	return "user"
}

func (t *MockProvisioner) ElevatedPassword() string {
	return "password"
}

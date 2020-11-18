package vagrant

import (
	"context"
	"testing"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
)

func TestStepSSHConfig_Impl(t *testing.T) {
	var raw interface{}
	raw = new(StepSSHConfig)
	if _, ok := raw.(multistep.Step); !ok {
		t.Fatalf("initialize should be a step")
	}
}

func TestPrepStepSSHConfig_GlobalID(t *testing.T) {
	driver := &MockVagrantDriver{}
	config := &Config{}
	state := new(multistep.BasicStateBag)
	state.Put("driver", driver)
	state.Put("config", config)

	step := StepSSHConfig{
		GlobalID: "adsfadf",
	}
	_ = step.Run(context.Background(), state)
	if driver.GlobalID != "adsfadf" {
		t.Fatalf("Should have called SSHConfig with GlobalID asdfasdf")
	}
}

func TestPrepStepSSHConfig_NoGlobalID(t *testing.T) {
	driver := &MockVagrantDriver{}
	config := &Config{}
	state := new(multistep.BasicStateBag)
	state.Put("driver", driver)
	state.Put("config", config)

	step := StepSSHConfig{}
	_ = step.Run(context.Background(), state)
	if driver.GlobalID != "source" {
		t.Fatalf("Should have called SSHConfig with GlobalID source")
	}
}

func TestPrepStepSSHConfig_SpacesInPath(t *testing.T) {
	driver := &MockVagrantDriver{}
	driver.ReturnSSHConfig = &VagrantSSHConfig{
		Hostname:               "127.0.0.1",
		User:                   "vagrant",
		Port:                   "2222",
		UserKnownHostsFile:     "/dev/null",
		StrictHostKeyChecking:  false,
		PasswordAuthentication: false,
		IdentityFile:           "\"/path with spaces/insecure_private_key\"",
		IdentitiesOnly:         true,
		LogLevel:               "FATAL"}

	config := &Config{}
	state := new(multistep.BasicStateBag)
	state.Put("driver", driver)
	state.Put("config", config)

	step := StepSSHConfig{}
	_ = step.Run(context.Background(), state)
	expected := "/path with spaces/insecure_private_key"
	if config.Comm.SSHPrivateKeyFile != expected {
		t.Fatalf("Bad config private key. Recieved: %s; expected: %s.", config.Comm.SSHPrivateKeyFile, expected)
	}
}

func TestPrepStepSSHConfig_NoSpacesInPath(t *testing.T) {
	driver := &MockVagrantDriver{}
	driver.ReturnSSHConfig = &VagrantSSHConfig{
		Hostname:               "127.0.0.1",
		User:                   "vagrant",
		Port:                   "2222",
		UserKnownHostsFile:     "/dev/null",
		StrictHostKeyChecking:  false,
		PasswordAuthentication: false,
		IdentityFile:           "/path/without/spaces/insecure_private_key",
		IdentitiesOnly:         true,
		LogLevel:               "FATAL"}

	config := &Config{}
	state := new(multistep.BasicStateBag)
	state.Put("driver", driver)
	state.Put("config", config)

	step := StepSSHConfig{}
	_ = step.Run(context.Background(), state)
	expected := "/path/without/spaces/insecure_private_key"
	if config.Comm.SSHPrivateKeyFile != expected {
		t.Fatalf("Bad config private key. Recieved: %s; expected: %s.", config.Comm.SSHPrivateKeyFile, expected)
	}
}

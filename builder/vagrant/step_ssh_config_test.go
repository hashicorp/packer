package vagrant

import (
	"context"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
)

func TestStepSSHConfig_Impl(t *testing.T) {
	var raw interface{}
	raw = new(StepSSHConfig)
	if _, ok := raw.(multistep.Step); !ok {
		t.Fatalf("initialize should be a step")
	}
}

func TestPrepStepSSHConfig_sshOverrides(t *testing.T) {
	type testcase struct {
		name              string
		inputSSHConfig    communicator.SSH
		expectedSSHConfig communicator.SSH
	}
	tcs := []testcase{
		{
			// defaults to overriding with the ssh config from vagrant\
			name:           "default",
			inputSSHConfig: communicator.SSH{},
			expectedSSHConfig: communicator.SSH{
				SSHHost:     "127.0.0.1",
				SSHPort:     2222,
				SSHUsername: "vagrant",
				SSHPassword: "",
			},
		},
		{
			// respects SSH host and port overrides independent of credential
			// overrides
			name: "host_override",
			inputSSHConfig: communicator.SSH{
				SSHHost: "123.45.67.8",
				SSHPort: 1234,
			},
			expectedSSHConfig: communicator.SSH{
				SSHHost:     "123.45.67.8",
				SSHPort:     1234,
				SSHUsername: "vagrant",
				SSHPassword: "",
			},
		},
		{
			// respects credential overrides
			name: "credential_override",
			inputSSHConfig: communicator.SSH{
				SSHUsername: "megan",
				SSHPassword: "SoSecure",
			},
			expectedSSHConfig: communicator.SSH{
				SSHHost:     "127.0.0.1",
				SSHPort:     2222,
				SSHUsername: "megan",
				SSHPassword: "SoSecure",
			},
		},
	}
	for _, tc := range tcs {
		driver := &MockVagrantDriver{}
		config := &Config{
			Comm: communicator.Config{
				SSH: tc.inputSSHConfig,
			},
		}
		state := new(multistep.BasicStateBag)
		state.Put("driver", driver)
		state.Put("config", config)

		step := StepSSHConfig{}
		_ = step.Run(context.Background(), state)

		if config.Comm.SSHHost != tc.expectedSSHConfig.SSHHost {
			t.Fatalf("unexpected sshconfig host: name: %s, recieved %s", tc.name, config.Comm.SSHHost)
		}
		if config.Comm.SSHPort != tc.expectedSSHConfig.SSHPort {
			t.Fatalf("unexpected sshconfig port: name: %s, recieved %d", tc.name, config.Comm.SSHPort)
		}
		if config.Comm.SSHUsername != tc.expectedSSHConfig.SSHUsername {
			t.Fatalf("unexpected sshconfig SSHUsername: name: %s, recieved %s", tc.name, config.Comm.SSHUsername)
		}
		if config.Comm.SSHPassword != tc.expectedSSHConfig.SSHPassword {
			t.Fatalf("unexpected sshconfig SSHUsername: name: %s, recieved %s", tc.name, config.Comm.SSHPassword)
		}
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

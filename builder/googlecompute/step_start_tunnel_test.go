package googlecompute

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"testing"

	"github.com/hashicorp/packer/helper/communicator"
)

type MockTunnelDriver struct {
	StopTunnelCalled   bool
	StartTunnelCalled  bool
	StartTunnelTimeout int
}

func (m *MockTunnelDriver) StopTunnel() {
	m.StopTunnelCalled = true
}

func (m *MockTunnelDriver) StartTunnel(_ context.Context, _ string, timeout int) error {
	m.StartTunnelCalled = true
	m.StartTunnelTimeout = timeout
	return nil
}

func getTestStepStartTunnel() *StepStartTunnel {
	return &StepStartTunnel{
		IAPConf: &IAPConfig{
			IAP:              true,
			IAPLocalhostPort: 0,
			IAPHashBang:      "/bin/bash",
			IAPExt:           "",
		},
		CommConf: &communicator.Config{
			SSH: communicator.SSH{
				SSHPort: 1234,
			},
		},
		AccountFile: "/path/to/account_file.json",
		ProjectId:   "fake-project-123",
	}
}

func TestStepStartTunnel_CreateTempScript(t *testing.T) {
	s := getTestStepStartTunnel()

	args := []string{"compute", "start-iap-tunnel", "fakeinstance-12345",
		"1234", "--local-host-port=localhost:8774", "--zone", "us-central-b"}

	scriptPath, err := s.createTempGcloudScript(args)
	if err != nil {
		t.Fatalf("Shouldn't have error building script file.")
	}
	defer os.Remove(scriptPath)

	f, err := ioutil.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("couldn't read created inventoryfile: %s", err)
	}

	expected := `#!/bin/bash

gcloud auth activate-service-account --key-file='/path/to/account_file.json'
gcloud config set project fake-project-123
gcloud compute start-iap-tunnel fakeinstance-12345 1234 --local-host-port=localhost:8774 --zone us-central-b
`
	if runtime.GOOS == "windows" {
		// in real life you'd not be passing a HashBang here, but GIGO.
		expected = `#!/bin/bash

call gcloud auth activate-service-account --key-file "/path/to/account_file.json"
call gcloud config set project fake-project-123
call gcloud compute start-iap-tunnel fakeinstance-12345 1234 --local-host-port=localhost:8774 --zone us-central-b
`
	}
	if fmt.Sprintf("%s", f) != expected {
		t.Fatalf("script didn't match expected:\n\n expected: \n%s\n; recieved: \n%s\n", expected, f)
	}
}

func TestStepStartTunnel_Cleanup(t *testing.T) {
	// Check IAP true
	s := getTestStepStartTunnel()
	td := &MockTunnelDriver{}
	s.tunnelDriver = td

	state := testState(t)
	s.Cleanup(state)

	if !td.StopTunnelCalled {
		t.Fatalf("Should have called StopTunnel, since IAP is true")
	}

	// Check IAP false
	s = getTestStepStartTunnel()
	td = &MockTunnelDriver{}
	s.tunnelDriver = td

	s.IAPConf.IAP = false

	s.Cleanup(state)

	if td.StopTunnelCalled {
		t.Fatalf("Should not have called StopTunnel, since IAP is false")
	}
}

func TestStepStartTunnel_ConfigurePort_port_set_by_user(t *testing.T) {
	s := getTestStepStartTunnel()
	s.IAPConf.IAPLocalhostPort = 8447

	ctx := context.TODO()
	err := s.ConfigureLocalHostPort(ctx)
	if err != nil {
		t.Fatalf("Shouldn't have error detecting port")
	}
	if s.IAPConf.IAPLocalhostPort != 8447 {
		t.Fatalf("Shouldn't have found new port; one was configured.")
	}
}

func TestStepStartTunnel_ConfigurePort_port_not_set_by_user(t *testing.T) {
	s := getTestStepStartTunnel()
	s.IAPConf.IAPLocalhostPort = 0

	ctx := context.TODO()
	err := s.ConfigureLocalHostPort(ctx)
	if err != nil {
		t.Fatalf("Shouldn't have error detecting port")
	}
	if s.IAPConf.IAPLocalhostPort == 0 {
		t.Fatalf("Should have found new port; none was configured.")
	}
}

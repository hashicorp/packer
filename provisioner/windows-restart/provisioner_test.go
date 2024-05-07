// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package restart

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

func testConfig() map[string]interface{} {
	return map[string]interface{}{}
}

func TestProvisioner_Impl(t *testing.T) {
	var raw interface{}
	raw = &Provisioner{}
	if _, ok := raw.(packersdk.Provisioner); !ok {
		t.Fatalf("must be a Provisioner")
	}
}

func TestProvisionerPrepare_Defaults(t *testing.T) {
	var p Provisioner
	config := testConfig()

	err := p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if p.config.RestartTimeout != 5*time.Minute {
		t.Errorf("unexpected restart timeout: %s", p.config.RestartTimeout)
	}

	if p.config.RestartCommand != "shutdown /r /f /t 0 /c \"packer restart\"" {
		t.Errorf("unexpected restart command: %s", p.config.RestartCommand)
	}
}

func TestProvisionerPrepare_ConfigRetryTimeout(t *testing.T) {
	var p Provisioner
	config := testConfig()
	config["restart_timeout"] = "1m"

	err := p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if p.config.RestartTimeout != 1*time.Minute {
		t.Errorf("unexpected restart timeout: %s", p.config.RestartTimeout)
	}
}

func TestProvisionerPrepare_ConfigErrors(t *testing.T) {
	var p Provisioner
	config := testConfig()
	config["restart_timeout"] = "m"

	err := p.Prepare(config)
	if err == nil {
		t.Fatal("Expected error parsing restart_timeout but did not receive one.")
	}
}

func TestProvisionerPrepare_InvalidKey(t *testing.T) {
	var p Provisioner
	config := testConfig()

	// Add a random key
	config["i_should_not_be_valid"] = true
	err := p.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}
}

func testUi() *packersdk.BasicUi {
	return &packersdk.BasicUi{
		Reader:      new(bytes.Buffer),
		Writer:      new(bytes.Buffer),
		ErrorWriter: new(bytes.Buffer),
	}
}

func TestProvisionerProvision_Success(t *testing.T) {
	config := testConfig()

	// Defaults provided by Packer
	ui := testUi()
	p := new(Provisioner)

	// Defaults provided by Packer
	comm := new(packersdk.MockCommunicator)
	p.Prepare(config)
	waitForCommunicatorOld := waitForCommunicator
	waitForCommunicator = func(context.Context, *Provisioner) error {
		return nil
	}
	waitForRestartOld := waitForRestart
	waitForRestart = func(context.Context, *Provisioner, packersdk.Communicator) error {
		return nil
	}
	err := p.Provision(context.Background(), ui, comm, make(map[string]interface{}))
	if err != nil {
		t.Fatal("should not have error")
	}

	expectedCommand := DefaultRestartCommand

	// Should run the command without alteration
	if comm.StartCmd.Command != expectedCommand {
		t.Fatalf("Expect command to be: %s, got %s", expectedCommand, comm.StartCmd.Command)
	}
	// Set this back!
	waitForCommunicator = waitForCommunicatorOld
	waitForRestart = waitForRestartOld
}

func TestProvisionerProvision_CustomCommand(t *testing.T) {
	config := testConfig()

	// Defaults provided by Packer
	ui := testUi()
	p := new(Provisioner)
	expectedCommand := "specialrestart.exe -NOW"
	config["restart_command"] = expectedCommand

	// Defaults provided by Packer
	comm := new(packersdk.MockCommunicator)
	p.Prepare(config)
	waitForCommunicatorOld := waitForCommunicator
	waitForCommunicator = func(context.Context, *Provisioner) error {
		return nil
	}
	waitForRestartOld := waitForRestart
	waitForRestart = func(context.Context, *Provisioner, packersdk.Communicator) error {
		return nil
	}
	err := p.Provision(context.Background(), ui, comm, make(map[string]interface{}))
	if err != nil {
		t.Fatal("should not have error")
	}

	// Should run the command without alteration
	if comm.StartCmd.Command != expectedCommand {
		t.Fatalf("Expect command to be: %s, got %s", expectedCommand, comm.StartCmd.Command)
	}
	// Set this back!
	waitForCommunicator = waitForCommunicatorOld
	waitForRestart = waitForRestartOld
}

func TestProvisionerProvision_RestartCommandFail(t *testing.T) {
	config := testConfig()
	ui := testUi()
	p := new(Provisioner)
	comm := new(packersdk.MockCommunicator)
	comm.StartStderr = "WinRM terminated"
	comm.StartExitStatus = 1

	p.Prepare(config)
	err := p.Provision(context.Background(), ui, comm, make(map[string]interface{}))
	if err == nil {
		t.Fatal("should have error")
	}
}
func TestProvisionerProvision_WaitForRestartFail(t *testing.T) {
	config := testConfig()

	// Defaults provided by Packer
	ui := testUi()
	p := new(Provisioner)

	// Defaults provided by Packer
	comm := new(packersdk.MockCommunicator)
	p.Prepare(config)
	waitForCommunicatorOld := waitForCommunicator
	waitForCommunicator = func(context.Context, *Provisioner) error {
		return fmt.Errorf("Machine did not restart properly")
	}
	err := p.Provision(context.Background(), ui, comm, make(map[string]interface{}))
	if err == nil {
		t.Fatal("should have error")
	}

	// Set this back!
	waitForCommunicator = waitForCommunicatorOld
}

func TestProvision_waitForRestartTimeout(t *testing.T) {
	retryableSleep = 10 * time.Millisecond
	config := testConfig()
	config["restart_timeout"] = "1ms"
	ui := testUi()
	p := new(Provisioner)
	comm := new(packersdk.MockCommunicator)
	var err error

	p.Prepare(config)
	waitForCommunicatorOld := waitForCommunicator
	waitDone := make(chan bool)
	waitContinue := make(chan bool)

	// Block until cancel comes through
	waitForCommunicator = func(context.Context, *Provisioner) error {
		for {
			select {
			case <-waitDone:
				waitContinue <- true
			}
		}
	}

	go func() {
		err = p.Provision(context.Background(), ui, comm, make(map[string]interface{}))
		waitDone <- true
	}()
	<-waitContinue

	if err == nil {
		t.Fatal("should not have error")
	}

	// Set this back!
	waitForCommunicator = waitForCommunicatorOld

}

func TestProvision_waitForCommunicator(t *testing.T) {
	config := testConfig()

	// Defaults provided by Packer
	ui := testUi()
	p := new(Provisioner)

	// Defaults provided by Packer
	comm := new(packersdk.MockCommunicator)
	p.comm = comm
	p.ui = ui
	comm.StartStderr = "WinRM terminated"
	comm.StartStdout = "WIN-V4CEJ7MC5SN restarted."
	comm.StartExitStatus = 1
	p.Prepare(config)
	err := waitForCommunicator(context.Background(), p)

	if err != nil {
		t.Fatalf("should not have error, got: %s", err.Error())
	}

	expectedCommand := DefaultRestartCheckCommand

	// Should run the command without alteration
	if comm.StartCmd.Command != expectedCommand {
		t.Fatalf("Expect command to be: %s, got %s", expectedCommand, comm.StartCmd.Command)
	}
}

func TestProvision_waitForCommunicatorWithCancel(t *testing.T) {
	config := testConfig()

	// Defaults provided by Packer
	ui := testUi()
	p := new(Provisioner)

	// Defaults provided by Packer
	comm := new(packersdk.MockCommunicator)
	p.comm = comm
	p.ui = ui
	retryableSleep = 5 * time.Second
	p.cancel = make(chan struct{})
	var err error

	ctx, cancel := context.WithCancel(context.Background())

	comm.StartStderr = "WinRM terminated"
	comm.StartExitStatus = 1 // Always fail
	p.Prepare(config)

	// Run 2 goroutines;
	//  1st to call waitForCommunicator (that will always fail)
	//  2nd to cancel the operation
	waitStart := make(chan bool)
	waitDone := make(chan bool)
	go func() {
		waitStart <- true
		err = waitForCommunicator(ctx, p)
		waitDone <- true
	}()

	go func() {
		time.Sleep(10 * time.Millisecond)
		<-waitStart
		cancel()
	}()
	<-waitDone

	// Expect a Cancel error
	if err == nil {
		t.Fatalf("Should have err")
	}
}

func TestProvision_Cancel(t *testing.T) {
	config := testConfig()

	// Defaults provided by Packer
	ui := testUi()
	p := new(Provisioner)

	comm := new(packersdk.MockCommunicator)
	p.Prepare(config)
	done := make(chan error)

	topCtx, cancelTopCtx := context.WithCancel(context.Background())

	// Block until cancel comes through
	waitForCommunicator = func(ctx context.Context, p *Provisioner) error {
		cancelTopCtx()
		<-ctx.Done()
		return ctx.Err()
	}

	// Create two go routines to provision and cancel in parallel
	// Provision will block until cancel happens
	go func() {
		done <- p.Provision(topCtx, ui, comm, make(map[string]interface{}))
	}()

	// Expect interrupt error
	if err := <-done; err == nil {
		t.Fatal("should have error")
	}
}

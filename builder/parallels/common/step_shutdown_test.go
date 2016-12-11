package common

import (
	"testing"
	"time"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

func TestStepShutdown_impl(t *testing.T) {
	var _ multistep.Step = new(StepShutdown)
}

func TestStepShutdown_noShutdownCommand(t *testing.T) {
	state := testState(t)
	step := new(StepShutdown)

	comm := new(packer.MockCommunicator)
	state.Put("communicator", comm)
	state.Put("vmName", "foo")

	driver := state.Get("driver").(*DriverMock)

	// Test the run
	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	// Test that Stop was just called
	if driver.StopName != "foo" {
		t.Fatal("should call stop")
	}
	if comm.StartCalled {
		t.Fatal("comm start should not be called")
	}
}

func TestStepShutdown_shutdownCommand(t *testing.T) {
	state := testState(t)
	step := new(StepShutdown)
	step.Command = "poweroff"
	step.Timeout = 1 * time.Second

	comm := new(packer.MockCommunicator)
	state.Put("communicator", comm)
	state.Put("vmName", "foo")

	driver := state.Get("driver").(*DriverMock)
	driver.IsRunningReturn = true

	go func() {
		time.Sleep(10 * time.Millisecond)
		driver.Lock()
		defer driver.Unlock()
		driver.IsRunningReturn = false
	}()

	// Test the run
	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	// Test that Stop was just called
	if driver.StopName != "" {
		t.Fatal("should not call stop")
	}
	if comm.StartCmd.Command != step.Command {
		t.Fatal("comm start should be called")
	}
}

func TestStepShutdown_shutdownTimeout(t *testing.T) {
	state := testState(t)
	step := new(StepShutdown)
	step.Command = "poweroff"
	step.Timeout = 1 * time.Second

	comm := new(packer.MockCommunicator)
	state.Put("communicator", comm)
	state.Put("vmName", "foo")

	driver := state.Get("driver").(*DriverMock)
	driver.IsRunningReturn = true

	go func() {
		time.Sleep(2 * time.Second)
		driver.Lock()
		defer driver.Unlock()
		driver.IsRunningReturn = false
	}()

	// Test the run
	if action := step.Run(state); action != multistep.ActionHalt {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); !ok {
		t.Fatal("should have error")
	}
}

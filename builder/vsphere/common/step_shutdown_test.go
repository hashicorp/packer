package common

import (
	"testing"
	"time"

	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
)

func testStepShutdownState(t *testing.T) multistep.StateBag {
	dir := testOutputDir(t)
	if err := dir.MkdirAll(); err != nil {
		t.Fatalf("err: %s", err)
	}

	state := testState(t)
	state.Put("communicator", new(packer.MockCommunicator))
	state.Put("dir", dir)
	return state
}

func TestStepShutdown_impl(t *testing.T) {
	var _ multistep.Step = new(StepShutdown)
}

func TestStepShutdown_command(t *testing.T) {
	state := testStepShutdownState(t)
	step := new(StepShutdown)
	step.Command = "foo"
	step.Timeout = 10 * time.Second
	step.Testing = true

	comm := state.Get("communicator").(*packer.MockCommunicator)
	driver := state.Get("driver").(*DriverMock)
	driver.IsRunningResult = true

	action := step.Run(state)
	// Test the run
	if action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	// Test the driver
	if driver.StopCalled {
		t.Fatal("stop should not be called")
	}

	if !comm.StartCalled {
		t.Fatal("start should be called")
	}
	if comm.StartCmd.Command != "foo" {
		t.Fatalf("bad: %#v", comm.StartCmd.Command)
	}
}

func TestStepShutdown_noCommand(t *testing.T) {
	state := testStepShutdownState(t)
	step := new(StepShutdown)

	comm := state.Get("communicator").(*packer.MockCommunicator)
	driver := state.Get("driver").(*DriverMock)

	// Test the run
	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	// Test the driver
	if !driver.StopCalled {
		t.Fatal("stop should be called")
	}

	if comm.StartCalled {
		t.Fatal("start should not be called")
	}
}

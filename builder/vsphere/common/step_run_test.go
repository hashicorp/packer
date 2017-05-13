package common

import (
	"testing"

	"github.com/mitchellh/multistep"
)

func TestStepRun_impl(t *testing.T) {
	var _ multistep.Step = new(StepRun)
}

func TestStepRun(t *testing.T) {
	state := testState(t)
	step := new(StepRun)

	driver := state.Get("driver").(*DriverMock)

	// Test the run
	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	// Test the driver
	if !driver.StartCalled {
		t.Fatal("start should be called")
	}

	// Test cleanup
	step.Cleanup(state)
	if !driver.StopCalled {
		t.Fatal("stop should be called if running")
	}
}

func TestStepRun_NotStarted(t *testing.T) {
	state := testState(t)
	step := new(StepRun)

	driver := state.Get("driver").(*DriverMock)

	// Test the run
	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	// Test the driver
	if !driver.StartCalled {
		t.Fatal("start should be called")
	}

	//Emulate error when starting the VM
	driver.StartCalled = false
	// Test cleanup
	step.Cleanup(state)
	if driver.StopCalled {
		t.Fatal("stop should not be called if not running")
	}
}

func TestStepRun_cleanupRunning(t *testing.T) {
	state := testState(t)
	step := new(StepRun)

	driver := state.Get("driver").(*DriverMock)

	// Test the run
	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	// Test the driver
	if !driver.StartCalled {
		t.Fatal("start should be called")
	}

	// Mark that it is running
	driver.IsRunningResult = true

	// Test cleanup
	step.Cleanup(state)
	if !driver.StopCalled {
		t.Fatal("stop should be called")
	}
}

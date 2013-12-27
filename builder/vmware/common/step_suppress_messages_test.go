package common

import (
	"testing"

	"github.com/mitchellh/multistep"
)

func TestStepSuppressMessages_impl(t *testing.T) {
	var _ multistep.Step = new(StepSuppressMessages)
}

func TestStepSuppressMessages(t *testing.T) {
	state := testState(t)
	step := new(StepSuppressMessages)

	state.Put("vmx_path", "foo")

	driver := state.Get("driver").(*DriverMock)

	// Test the run
	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	// Test the driver
	if !driver.SuppressMessagesCalled {
		t.Fatal("should've called")
	}
	if driver.SuppressMessagesPath != "foo" {
		t.Fatal("should call with right path")
	}
}

package common

import (
	"github.com/mitchellh/multistep"
	"testing"
)

func TestStepRegister_impl(t *testing.T) {
	var _ multistep.Step = new(StepRegister)
}

func TestStepRegister(t *testing.T) {
	state := testState(t)
	step := new(StepRegister)

	driver := new(DriverMock)
	state.Put("driver", driver)

	step.KeepRegistered = false

	// Test the run
	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	// cleanup
	step.Cleanup(state)
	if !driver.DestroyCalled {
		t.Fatal("destroy should be called")
	}
	if !driver.IsDestroyedCalled {
		t.Fatal("isdestroyed should be called")
	}
}
func TestStepRegister_WithoutUnregister(t *testing.T) {
	state := testState(t)
	step := new(StepRegister)

	driver := new(DriverMock)
	step.KeepRegistered = true

	state.Put("driver", driver)

	// cleanup
	step.Cleanup(state)
	if driver.DestroyCalled {
		t.Fatal("destroy should not be called")
	}
}

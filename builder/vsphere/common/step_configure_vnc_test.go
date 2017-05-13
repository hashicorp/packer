package common

import (
	"testing"

	"github.com/mitchellh/multistep"
)

func TestStepConfigureVNC_impl(t *testing.T) {
	var _ multistep.Step = new(StepConfigureVNC)
}

func TestStepConfigureVNC(t *testing.T) {
	state := testState(t)
	step := new(StepConfigureVNC)
	step.VNCPortMin = 5555
	step.VNCPortMax = 6666
	step.VNCDisablePassword = false

	driver := state.Get("driver").(*DriverMock)

	// Test the run
	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	// Test the driver
	if !driver.VNCEnableCalled {
		t.Fatal("VNCEnable should be called")
	}
}

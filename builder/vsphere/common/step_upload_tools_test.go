package common

import (
	"testing"

	"github.com/mitchellh/multistep"
)

func TestStepUploadTools_impl(t *testing.T) {
	var _ multistep.Step = new(StepUploadTools)
}

func TestStepUploadTools(t *testing.T) {
	state := testState(t)
	driver := new(DriverMock)
	state.Put("driver", driver)

	step := new(StepUploadTools)

	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	if !driver.ToolsInstallCalled {
		t.Fatal("upload tools should be called")
	}
	// Cleanup
	step.Cleanup(state)
}

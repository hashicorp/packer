package common

import (
	"github.com/mitchellh/multistep"
	"testing"
)

func TestStepExport_impl(t *testing.T) {
	var _ multistep.Step = new(StepExport)
}

func TestStepExport_skip(t *testing.T) {
	state := testState(t)
	driver := new(DriverMock)
	state.Put("driver", driver)

	step := new(StepExport)
	step.Format = "ovf"
	step.SkipExport = true

	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	if driver.ExportVirtualMachineCalled {
		t.Fatal("export should not be called")
	}
	// Cleanup
	step.Cleanup(state)
}

func TestStepExport(t *testing.T) {
	state := testState(t)
	driver := new(DriverMock)
	state.Put("driver", driver)

	step := new(StepExport)
	step.Format = "ovf"
	step.OutputPath = "foo-dir"
	step.SkipExport = false

	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	if !driver.ExportVirtualMachineCalled {
		t.Fatal("export should be called")
	}
	// Cleanup
	step.Cleanup(state)
}

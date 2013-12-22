package common

import (
	"github.com/mitchellh/multistep"
	"testing"
)

func TestStepExport_impl(t *testing.T) {
	var _ multistep.Step = new(StepExport)
}

func TestStepExport(t *testing.T) {
	state := testState(t)
	step := new(StepExport)

	state.Put("vmName", "foo")

	driver := state.Get("driver").(*DriverMock)

	// Test the run
	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	// Test output state
	if _, ok := state.GetOk("exportPath"); !ok {
		t.Fatal("should set exportPath")
	}

	// Test driver
	if len(driver.VBoxManageCalls) != 2 {
		t.Fatal("should call vboxmanage")
	}
	if driver.VBoxManageCalls[0][0] != "modifyvm" {
		t.Fatal("bad")
	}
	if driver.VBoxManageCalls[1][0] != "export" {
		t.Fatal("bad")
	}
}

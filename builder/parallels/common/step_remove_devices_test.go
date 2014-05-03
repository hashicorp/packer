package common

import (
	"github.com/mitchellh/multistep"
	"testing"
)

func TestStepRemoveDevices_impl(t *testing.T) {
	var _ multistep.Step = new(StepRemoveDevices)
}

func TestStepRemoveDevices(t *testing.T) {
	state := testState(t)
	step := new(StepRemoveDevices)

	state.Put("vmName", "foo")

	driver := state.Get("driver").(*DriverMock)

	// Test the run
	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	// Test that ISO was removed
	if len(driver.PrlctlCalls) != 0 {
		t.Fatalf("bad: %#v", driver.PrlctlCalls)
	}
}

func TestStepRemoveDevices_attachedIso(t *testing.T) {
	state := testState(t)
	step := new(StepRemoveDevices)

	state.Put("attachedIso", true)
	state.Put("vmName", "foo")

	driver := state.Get("driver").(*DriverMock)

	// Test the run
	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	// Test that ISO was detached
	if len(driver.PrlctlCalls) != 1 {
		t.Fatalf("bad: %#v", driver.PrlctlCalls)
	}
	if driver.PrlctlCalls[0][2] != "--device-set" {
		t.Fatalf("bad: %#v", driver.PrlctlCalls)
	}
	if driver.PrlctlCalls[0][3] != "cdrom0" {
		t.Fatalf("bad: %#v", driver.PrlctlCalls)
	}
	if driver.PrlctlCalls[0][5] != "Default CD/DVD-ROM" {
		t.Fatalf("bad: %#v", driver.PrlctlCalls)
	}
}

func TestStepRemoveDevices_floppyPath(t *testing.T) {
	state := testState(t)
	step := new(StepRemoveDevices)

	state.Put("floppy_path", "foo")
	state.Put("vmName", "foo")

	driver := state.Get("driver").(*DriverMock)

	// Test the run
	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	// Test that both were removed
	if len(driver.PrlctlCalls) != 1 {
		t.Fatalf("bad: %#v", driver.PrlctlCalls)
	}
	if driver.PrlctlCalls[0][2] != "--device-del" {
		t.Fatalf("bad: %#v", driver.PrlctlCalls)
	}
	if driver.PrlctlCalls[0][3] != "fdd0" {
		t.Fatalf("bad: %#v", driver.PrlctlCalls)
	}
}

package common

import (
	"testing"

	"github.com/mitchellh/multistep"
)

func TestStepCompactDisk_impl(t *testing.T) {
	var _ multistep.Step = new(StepCompactDisk)
}

func TestStepCompactDisk(t *testing.T) {
	state := testState(t)
	step := new(StepCompactDisk)

	full_disk_paths := make([]string, 0)
	full_disk_paths = append(full_disk_paths, "foo", "bar")
	state.Put("full_disk_paths", full_disk_paths)

	driver := state.Get("driver").(*DriverMock)

	// Test the run
	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	// Test the driver
	if !driver.CompactDiskCalled {
		t.Fatal("should've called")
	}
	if driver.CompactDiskPath[0] != "foo" && driver.CompactDiskPath[1] != "bar" {
		t.Fatal("should call with right path")
	}
}

func TestStepCompactDisk_skip(t *testing.T) {
	state := testState(t)
	step := new(StepCompactDisk)
	step.Skip = true

	full_disk_paths := make([]string, 0)
	full_disk_paths = append(full_disk_paths, "foo", "bar")
	state.Put("full_disk_paths", full_disk_paths)

	driver := state.Get("driver").(*DriverMock)

	// Test the run
	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	// Test the driver
	if driver.CompactDiskCalled {
		t.Fatal("should not have called")
	}
}

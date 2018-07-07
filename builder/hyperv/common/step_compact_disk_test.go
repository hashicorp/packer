package common

import (
	"context"
	"testing"

	"github.com/hashicorp/packer/helper/multistep"
)

func TestStepCompactDisk_impl(t *testing.T) {
	var _ multistep.Step = new(StepCompactDisk)
}

func TestStepCompactDisk(t *testing.T) {
	state := testState(t)
	step := new(StepCompactDisk)

	// Set up the path to the tmp directory used to store VM files
	tmpPath := "foopath"
	state.Put("packerTempDir", tmpPath)

	driver := state.Get("driver").(*DriverMock)

	// Test the run
	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("Bad action: %v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("Should NOT have error")
	}

	// Test the driver
	if !driver.CompactDisks_Called {
		t.Fatal("Should have called CompactDisks")
	}
	if driver.CompactDisks_Path != tmpPath {
		t.Fatalf("Should call with correct path. Got: %s Wanted: %s", driver.CompactDisks_Path, tmpPath)
	}
}

func TestStepCompactDisk_skip(t *testing.T) {
	state := testState(t)
	step := new(StepCompactDisk)
	step.SkipCompaction = true

	// Set up the path to the tmp directory used to store VM files
	state.Put("packerTempDir", "foopath")

	driver := state.Get("driver").(*DriverMock)

	// Test the run
	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("Bad action: %v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatalf("Should NOT have error")
	}

	// Test the driver
	if driver.CompactDisks_Called {
		t.Fatal("Should NOT have called CompactDisks")
	}
}

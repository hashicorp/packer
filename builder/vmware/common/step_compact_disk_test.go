package common

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/hashicorp/packer/helper/multistep"
)

func TestStepCompactDisk_impl(t *testing.T) {
	var _ multistep.Step = new(StepCompactDisk)
}

func TestStepCompactDisk(t *testing.T) {
	state := testState(t)
	step := new(StepCompactDisk)

	// Create a fake vmdk file for disk file size operations
	diskFile, err := ioutil.TempFile("", "disk.vmdk")
	if err != nil {
		t.Fatalf("Error creating fake vmdk file: %s", err)
	}

	diskFullPath := diskFile.Name()
	defer os.Remove(diskFullPath)

	content := []byte("I am the fake vmdk's contents")
	if _, err := diskFile.Write(content); err != nil {
		t.Fatalf("Error writing to fake vmdk file: %s", err)
	}
	if err := diskFile.Close(); err != nil {
		t.Fatalf("Error closing fake vmdk file: %s", err)
	}

	// Set up required state
	state.Put("disk_full_paths", []string{diskFullPath})

	driver := state.Get("driver").(*DriverMock)

	// Test the run
	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	// Test the driver
	if !driver.CompactDiskCalled {
		t.Fatal("should've called")
	}
	if driver.CompactDiskPath != diskFullPath {
		t.Fatal("should call with right path")
	}
}

func TestStepCompactDisk_skip(t *testing.T) {
	state := testState(t)
	step := new(StepCompactDisk)
	step.Skip = true

	diskFullPaths := []string{"foo"}
	state.Put("disk_full_paths", diskFullPaths)

	driver := state.Get("driver").(*DriverMock)

	// Test the run
	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
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

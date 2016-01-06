package common

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/mitchellh/multistep"
)

func TestStepCompactDisk_impl(t *testing.T) {
	var _ multistep.Step = new(StepCompactDisk)
}

func TestStepCompactDisk(t *testing.T) {
	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	tf.Close()
	defer os.Remove(tf.Name())

	state := testState(t)
	step := new(StepCompactDisk)

	state.Put("vmName", "foo")

	driver := state.Get("driver").(*DriverMock)

	// Mock results
	driver.DiskPathResult = tf.Name()

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

	path, _ := driver.DiskPath("foo")
	if path != tf.Name() {
		t.Fatal("should call with right path")
	}
}

func TestStepCompactDisk_skip(t *testing.T) {
	state := testState(t)
	step := new(StepCompactDisk)
	step.Skip = true

	state.Put("vmName", "foo")

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

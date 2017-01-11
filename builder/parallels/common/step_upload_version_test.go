package common

import (
	"testing"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

func TestStepUploadVersion_impl(t *testing.T) {
	var _ multistep.Step = new(StepUploadVersion)
}

func TestStepUploadVersion(t *testing.T) {
	state := testState(t)
	step := new(StepUploadVersion)
	step.Path = "foopath"

	comm := new(packer.MockCommunicator)
	state.Put("communicator", comm)

	driver := state.Get("driver").(*DriverMock)
	driver.VersionResult = "foo"

	// Test the run
	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	// Verify
	if comm.UploadPath != "foopath" {
		t.Fatalf("bad: %#v", comm.UploadPath)
	}
	if comm.UploadData != "foo" {
		t.Fatalf("upload data bad: %#v", comm.UploadData)
	}
}

func TestStepUploadVersion_noPath(t *testing.T) {
	state := testState(t)
	step := new(StepUploadVersion)
	step.Path = ""

	comm := new(packer.MockCommunicator)
	state.Put("communicator", comm)

	// Test the run
	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	// Verify
	if comm.UploadCalled {
		t.Fatal("bad")
	}
}

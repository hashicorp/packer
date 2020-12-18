package common

import (
	"context"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

func TestStepUploadVersion_impl(t *testing.T) {
	var _ multistep.Step = new(StepUploadVersion)
}

func TestStepUploadVersion(t *testing.T) {
	state := testState(t)
	step := new(StepUploadVersion)
	step.Path = "foopath"

	comm := new(packersdk.MockCommunicator)
	state.Put("communicator", comm)

	driver := state.Get("driver").(*DriverMock)
	driver.VersionResult = "foo"

	// Test the run
	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
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

	comm := new(packersdk.MockCommunicator)
	state.Put("communicator", comm)

	// Test the run
	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
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

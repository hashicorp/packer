package common

import (
	"testing"

	"github.com/mitchellh/multistep"
)

func TestStepRemoteUpload_impl(t *testing.T) {
	var _ multistep.Step = new(StepRemoteUpload)
}

func TestStepRemoteUpload(t *testing.T) {
	state := testState(t)
	driver := new(DriverMock)
	state.Put("driver", driver)
	state.Put("foo_path", "foo")

	step := new(StepRemoteUpload)
	step.Key = "foo_path"

	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	if !driver.UploadCalled {
		t.Fatal("upload tools should be called")
	}

	remotepath := state.Get("foo_path")
	if remotepath != "/datastore/foo" {
		t.Fatalf("wrong remotepath: %s", remotepath)
	}

	// Cleanup
	step.Cleanup(state)
}

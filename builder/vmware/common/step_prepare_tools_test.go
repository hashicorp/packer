package common

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/mitchellh/multistep"
)

func TestStepPrepareTools_impl(t *testing.T) {
	var _ multistep.Step = new(StepPrepareTools)
}

func TestStepPrepareTools(t *testing.T) {
	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	tf.Close()
	defer os.Remove(tf.Name())

	state := testState(t)
	step := &StepPrepareTools{
		RemoteType:        "",
		ToolsUploadFlavor: "foo",
	}

	driver := state.Get("driver").(*DriverMock)

	// Mock results
	driver.ToolsIsoPathResult = tf.Name()

	// Test the run
	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	// Test the driver
	if !driver.ToolsIsoPathCalled {
		t.Fatal("tools iso path should be called")
	}
	if driver.ToolsIsoPathFlavor != "foo" {
		t.Fatalf("bad: %#v", driver.ToolsIsoPathFlavor)
	}

	// Test the resulting state
	path, ok := state.GetOk("tools_upload_source")
	if !ok {
		t.Fatal("should have tools_upload_source")
	}
	if path != tf.Name() {
		t.Fatalf("bad: %#v", path)
	}
}

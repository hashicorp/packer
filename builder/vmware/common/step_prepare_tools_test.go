package common

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
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
	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
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

func TestStepPrepareTools_esx5(t *testing.T) {
	state := testState(t)
	step := &StepPrepareTools{
		RemoteType:        "esx5",
		ToolsUploadFlavor: "foo",
	}

	driver := state.Get("driver").(*DriverMock)

	// Test the run
	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	// Test the driver
	if driver.ToolsIsoPathCalled {
		t.Fatal("tools iso path should NOT be called")
	}
}

func TestStepPrepareTools_nonExist(t *testing.T) {
	state := testState(t)
	step := &StepPrepareTools{
		RemoteType:        "",
		ToolsUploadFlavor: "foo",
	}

	driver := state.Get("driver").(*DriverMock)

	// Mock results
	driver.ToolsIsoPathResult = "foo"

	// Test the run
	if action := step.Run(context.Background(), state); action != multistep.ActionHalt {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); !ok {
		t.Fatal("should have error")
	}

	// Test the driver
	if !driver.ToolsIsoPathCalled {
		t.Fatal("tools iso path should be called")
	}
	if driver.ToolsIsoPathFlavor != "foo" {
		t.Fatalf("bad: %#v", driver.ToolsIsoPathFlavor)
	}

	// Test the resulting state
	if _, ok := state.GetOk("tools_upload_source"); ok {
		t.Fatal("should NOT have tools_upload_source")
	}
}

func TestStepPrepareTools_SourcePath(t *testing.T) {
	state := testState(t)
	step := &StepPrepareTools{
		RemoteType:      "",
		ToolsSourcePath: "/path/to/tool.iso",
	}

	driver := state.Get("driver").(*DriverMock)

	// Mock results
	driver.ToolsIsoPathResult = "foo"

	// Test the run
	if action := step.Run(context.Background(), state); action != multistep.ActionHalt {
		t.Fatalf("Should have failed when stat failed %#v", action)
	}
	if _, ok := state.GetOk("error"); !ok {
		t.Fatal("should have error")
	}

	// Test the driver
	if driver.ToolsIsoPathCalled {
		t.Fatal("tools iso path should not be called when ToolsSourcePath is set")
	}

	// Test the resulting state
	if _, ok := state.GetOk("tools_upload_source"); ok {
		t.Fatal("should NOT have tools_upload_source")
	}
}

func TestStepPrepareTools_SourcePath_exists(t *testing.T) {
	state := testState(t)
	step := &StepPrepareTools{
		RemoteType:      "",
		ToolsSourcePath: "./step_prepare_tools.go",
	}

	driver := state.Get("driver").(*DriverMock)

	// Mock results
	driver.ToolsIsoPathResult = "foo"

	// Test the run
	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("Step should succeed when stat succeeds: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	// Test the driver
	if driver.ToolsIsoPathCalled {
		t.Fatal("tools iso path should not be called when ToolsSourcePath is set")
	}

	// Test the resulting state
	if _, ok := state.GetOk("tools_upload_source"); !ok {
		t.Fatal("should have tools_upload_source")
	}
}

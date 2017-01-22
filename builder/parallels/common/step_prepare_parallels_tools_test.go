package common

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/mitchellh/multistep"
)

func TestStepPrepareParallelsTools_impl(t *testing.T) {
	var _ multistep.Step = new(StepPrepareParallelsTools)
}

func TestStepPrepareParallelsTools(t *testing.T) {
	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	tf.Close()
	defer os.Remove(tf.Name())

	state := testState(t)
	step := &StepPrepareParallelsTools{
		ParallelsToolsMode:   "",
		ParallelsToolsFlavor: "foo",
	}

	driver := state.Get("driver").(*DriverMock)

	// Mock results
	driver.ToolsISOPathResult = tf.Name()

	// Test the run
	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	// Test the driver
	if !driver.ToolsISOPathCalled {
		t.Fatal("tools iso path should be called")
	}
	if driver.ToolsISOPathFlavor != "foo" {
		t.Fatalf("bad: %#v", driver.ToolsISOPathFlavor)
	}

	// Test the resulting state
	path, ok := state.GetOk("parallels_tools_path")
	if !ok {
		t.Fatal("should have parallels_tools_path")
	}
	if path != tf.Name() {
		t.Fatalf("bad: %#v", path)
	}
}

func TestStepPrepareParallelsTools_disabled(t *testing.T) {
	state := testState(t)
	step := &StepPrepareParallelsTools{
		ParallelsToolsFlavor: "foo",
		ParallelsToolsMode:   ParallelsToolsModeDisable,
	}

	driver := state.Get("driver").(*DriverMock)

	// Test the run
	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	// Test the driver
	if driver.ToolsISOPathCalled {
		t.Fatal("tools ISO path should NOT be called")
	}
}

func TestStepPrepareParallelsTools_nonExist(t *testing.T) {
	state := testState(t)
	step := &StepPrepareParallelsTools{
		ParallelsToolsFlavor: "foo",
		ParallelsToolsMode:   "",
	}

	driver := state.Get("driver").(*DriverMock)

	// Mock results
	driver.ToolsISOPathResult = "foo"

	// Test the run
	if action := step.Run(state); action != multistep.ActionHalt {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); !ok {
		t.Fatal("should have error")
	}

	// Test the driver
	if !driver.ToolsISOPathCalled {
		t.Fatal("tools iso path should be called")
	}
	if driver.ToolsISOPathFlavor != "foo" {
		t.Fatalf("bad: %#v", driver.ToolsISOPathFlavor)
	}

	// Test the resulting state
	if _, ok := state.GetOk("parallels_tools_path"); ok {
		t.Fatal("should NOT have parallels_tools_path")
	}
}

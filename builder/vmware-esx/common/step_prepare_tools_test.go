package common

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/hashicorp/packer/helper/multistep"
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
	if driver.ToolsIsoPathCalled {
		t.Fatal("tools iso path should NOT be called")
	}
	if driver.ToolsIsoPathFlavor != "foo" {
		t.Fatalf("bad: %#v", driver.ToolsIsoPathFlavor)
	}
}

func TestStepPrepareTools_esx5(t *testing.T) {
	state := testState(t)
	step := &StepPrepareTools{
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
		ToolsUploadFlavor: "foo",
	}

	driver := state.Get("driver").(*DriverMock)

	// Mock results
	driver.ToolsIsoPathResult = "foo"

	// Test the run
	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); !ok {
		t.Fatal("should have error")
	}

	// Test the driver
	if driver.ToolsIsoPathCalled {
		t.Fatal("tools iso path should NOT be called")
	}
	if driver.ToolsIsoPathFlavor != "foo" {
		t.Fatalf("bad: %#v", driver.ToolsIsoPathFlavor)
	}
}

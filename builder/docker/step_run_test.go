package docker

import (
	"errors"
	"github.com/mitchellh/multistep"
	"testing"
)

func testStepRunState(t *testing.T) multistep.StateBag {
	state := testState(t)
	state.Put("temp_dir", "/foo")
	return state
}

func TestStepRun_impl(t *testing.T) {
	var _ multistep.Step = new(StepRun)
}

func TestStepRun(t *testing.T) {
	state := testStepRunState(t)
	step := new(StepRun)
	defer step.Cleanup(state)

	config := state.Get("config").(*Config)
	driver := state.Get("driver").(*MockDriver)
	driver.StartID = "foo"

	// run the step
	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}

	// verify we did the right thing
	if !driver.StartCalled {
		t.Fatal("should've called")
	}
	if driver.StartConfig.Image != config.Image {
		t.Fatalf("bad: %#v", driver.StartConfig.Image)
	}

	// verify the ID is saved
	idRaw, ok := state.GetOk("container_id")
	if !ok {
		t.Fatal("should've saved ID")
	}

	id := idRaw.(string)
	if id != "foo" {
		t.Fatalf("bad: %#v", id)
	}

	// Verify we haven't called stop yet
	if driver.StopCalled {
		t.Fatal("should not have stopped")
	}

	// Cleanup
	step.Cleanup(state)
	if !driver.StopCalled {
		t.Fatal("should've stopped")
	}
	if driver.StopID != id {
		t.Fatalf("bad: %#v", driver.StopID)
	}
}

func TestStepRun_error(t *testing.T) {
	state := testStepRunState(t)
	step := new(StepRun)
	defer step.Cleanup(state)

	driver := state.Get("driver").(*MockDriver)
	driver.StartError = errors.New("foo")

	// run the step
	if action := step.Run(state); action != multistep.ActionHalt {
		t.Fatalf("bad action: %#v", action)
	}

	// verify the ID is not saved
	if _, ok := state.GetOk("container_id"); ok {
		t.Fatal("shouldn't save container ID")
	}

	// Verify we haven't called stop yet
	if driver.StopCalled {
		t.Fatal("should not have stopped")
	}

	// Cleanup
	step.Cleanup(state)
	if driver.StopCalled {
		t.Fatal("should not have stopped")
	}
}

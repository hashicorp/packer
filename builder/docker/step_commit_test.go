package docker

import (
	"errors"
	"github.com/mitchellh/multistep"
	"testing"
)

func testStepCommitState(t *testing.T) multistep.StateBag {
	state := testState(t)
	state.Put("container_id", "foo")
	return state
}

func TestStepCommit_impl(t *testing.T) {
	var _ multistep.Step = new(StepCommit)
}

func TestStepCommit(t *testing.T) {
	state := testStepCommitState(t)
	step := new(StepCommit)
	defer step.Cleanup(state)

	driver := state.Get("driver").(*MockDriver)
	driver.CommitImageId = "bar"

	// run the step
	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}

	// verify we did the right thing
	if !driver.CommitCalled {
		t.Fatal("should've called")
	}

	// verify the ID is saved
	idRaw, ok := state.GetOk("image_id")
	if !ok {
		t.Fatal("should've saved ID")
	}

	id := idRaw.(string)
	if id != driver.CommitImageId {
		t.Fatalf("bad: %#v", id)
	}
}

func TestStepCommit_error(t *testing.T) {
	state := testStepCommitState(t)
	step := new(StepCommit)
	defer step.Cleanup(state)

	driver := state.Get("driver").(*MockDriver)
	driver.CommitErr = errors.New("foo")

	// run the step
	if action := step.Run(state); action != multistep.ActionHalt {
		t.Fatalf("bad action: %#v", action)
	}

	// verify the ID is not saved
	if _, ok := state.GetOk("image_id"); ok {
		t.Fatal("shouldn't save image ID")
	}
}

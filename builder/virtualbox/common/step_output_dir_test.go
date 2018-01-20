package common

import (
	"github.com/hashicorp/packer/helper/multistep"
	"io/ioutil"
	"os"
	"testing"

	"github.com/mitchellh/multistep"
)

func testStepOutputDir(t *testing.T) *StepOutputDir {
	td, err := ioutil.TempDir("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if err := os.RemoveAll(td); err != nil {
		t.Fatalf("err: %s", err)
	}

	return &StepOutputDir{Force: false, Path: td}
}

func TestStepOutputDir_impl(t *testing.T) {
	var _ multistep.Step = new(StepOutputDir)
}

func TestStepOutputDir(t *testing.T) {
	state := testState(t)
	step := testStepOutputDir(t)

	// Test the run
	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}
	if _, err := os.Stat(step.Path); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Test the cleanup
	step.Cleanup(state)
	if _, err := os.Stat(step.Path); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestStepOutputDir_exists(t *testing.T) {
	state := testState(t)
	step := testStepOutputDir(t)

	// Make the dir
	if err := os.MkdirAll(step.Path, 0755); err != nil {
		t.Fatalf("bad: %s", err)
	}

	// Test the run
	if action := step.Run(state); action != multistep.ActionHalt {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); !ok {
		t.Fatal("should have error")
	}

	// Test the cleanup
	step.Cleanup(state)
	if _, err := os.Stat(step.Path); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestStepOutputDir_cancelled(t *testing.T) {
	state := testState(t)
	step := testStepOutputDir(t)

	// Test the run
	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}
	if _, err := os.Stat(step.Path); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Mark
	state.Put(multistep.StateCancelled, true)

	// Test the cleanup
	step.Cleanup(state)
	if _, err := os.Stat(step.Path); err == nil {
		t.Fatal("should not exist")
	}
}

func TestStepOutputDir_halted(t *testing.T) {
	state := testState(t)
	step := testStepOutputDir(t)

	// Test the run
	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}
	if _, err := os.Stat(step.Path); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Mark
	state.Put(multistep.StateHalted, true)

	// Test the cleanup
	step.Cleanup(state)
	if _, err := os.Stat(step.Path); err == nil {
		t.Fatal("should not exist")
	}
}

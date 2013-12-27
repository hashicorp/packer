package common

import (
	"github.com/mitchellh/multistep"
	"io/ioutil"
	"os"
	"testing"
)

func testOutputDir(t *testing.T) *LocalOutputDir {
	td, err := ioutil.TempDir("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	os.RemoveAll(td)

	result := new(LocalOutputDir)
	result.SetOutputDir(td)
	return result
}

func TestStepOutputDir_impl(t *testing.T) {
	var _ multistep.Step = new(StepOutputDir)
}

func TestStepOutputDir(t *testing.T) {
	state := testState(t)
	step := new(StepOutputDir)

	dir := testOutputDir(t)
	state.Put("dir", dir)

	// Test the run
	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}
	if _, err := os.Stat(dir.dir); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Test the cleanup
	step.Cleanup(state)
	if _, err := os.Stat(dir.dir); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestStepOutputDir_existsNoForce(t *testing.T) {
	state := testState(t)
	step := new(StepOutputDir)

	dir := testOutputDir(t)
	state.Put("dir", dir)

	// Make sure the dir exists
	if err := os.MkdirAll(dir.dir, 0755); err != nil {
		t.Fatalf("err: %s", err)
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
	if _, err := os.Stat(dir.dir); err != nil {
		t.Fatal("should not delete dir")
	}
}

func TestStepOutputDir_existsForce(t *testing.T) {
	state := testState(t)
	step := new(StepOutputDir)
	step.Force = true

	dir := testOutputDir(t)
	state.Put("dir", dir)

	// Make sure the dir exists
	if err := os.MkdirAll(dir.dir, 0755); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Test the run
	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}
	if _, err := os.Stat(dir.dir); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestStepOutputDir_cancel(t *testing.T) {
	state := testState(t)
	step := new(StepOutputDir)

	dir := testOutputDir(t)
	state.Put("dir", dir)

	// Test the run
	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}
	if _, err := os.Stat(dir.dir); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Test cancel/halt
	state.Put(multistep.StateCancelled, true)
	step.Cleanup(state)
	if _, err := os.Stat(dir.dir); err == nil {
		t.Fatal("directory should not exist")
	}
}

func TestStepOutputDir_halt(t *testing.T) {
	state := testState(t)
	step := new(StepOutputDir)

	dir := testOutputDir(t)
	state.Put("dir", dir)

	// Test the run
	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}
	if _, err := os.Stat(dir.dir); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Test cancel/halt
	state.Put(multistep.StateHalted, true)
	step.Cleanup(state)
	if _, err := os.Stat(dir.dir); err == nil {
		t.Fatal("directory should not exist")
	}
}

package common

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/mitchellh/multistep"
)

func TestStepCleanFiles_impl(t *testing.T) {
	var _ multistep.Step = new(StepCleanFiles)
}

func TestStepCleanFiles(t *testing.T) {
	state := testState(t)
	step := new(StepCleanFiles)

	td, err := ioutil.TempDir("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.RemoveAll(td)

	if err := ioutil.WriteFile(filepath.Join(td, "file.vmx"), nil, 0644); err != nil {
		t.Fatalf("err: %s", err)
	}

	if err := ioutil.WriteFile(filepath.Join(td, "file.txt"), nil, 0644); err != nil {
		t.Fatalf("err: %s", err)
	}

	dir := new(LocalOutputDir)
	dir.SetOutputDir(td)

	state.Put("dir", dir)

	// Test the run
	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	if _, err := os.Stat(filepath.Join(td, "file.vmx")); os.IsNotExist(err) {
		t.Fatal("should NOT have deleted file.vmx")
	}

	if _, err := os.Stat(filepath.Join(td, "file.txt")); err == nil {
		t.Fatal("should have deleted file.txt")
	}
}

func TestStepCleanFiles_skip(t *testing.T) {
	state := testState(t)
	step := new(StepCleanFiles)
	step.Skip = true

	td, err := ioutil.TempDir("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.RemoveAll(td)

	if err := ioutil.WriteFile(filepath.Join(td, "file.vmx"), nil, 0644); err != nil {
		t.Fatalf("err: %s", err)
	}

	if err := ioutil.WriteFile(filepath.Join(td, "file.txt"), nil, 0644); err != nil {
		t.Fatalf("err: %s", err)
	}

	dir := new(LocalOutputDir)
	dir.SetOutputDir(td)

	state.Put("dir", dir)

	// Test the run
	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	if _, err := os.Stat(filepath.Join(td, "file.vmx")); os.IsNotExist(err) {
		t.Fatal("should NOT have deleted file.vmx")
	}

	if _, err := os.Stat(filepath.Join(td, "file.txt")); os.IsNotExist(err) {
		t.Fatal("should NOT have deleted file.txt")
	}
}

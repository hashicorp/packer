package docker

import (
	"github.com/mitchellh/multistep"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestStepTempDir_impl(t *testing.T) {
	var _ multistep.Step = new(StepTempDir)
}

func testStepTempDir_impl(t *testing.T) string {
	state := testState(t)
	step := new(StepTempDir)
	defer step.Cleanup(state)

	// sanity test
	if _, ok := state.GetOk("temp_dir"); ok {
		t.Fatalf("temp_dir should not be in state yet")
	}

	// run the step
	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}

	// Verify that we got the temp dir
	dirRaw, ok := state.GetOk("temp_dir")
	if !ok {
		t.Fatalf("should've made temp_dir")
	}
	dir := dirRaw.(string)

	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Cleanup
	step.Cleanup(state)
	if _, err := os.Stat(dir); err == nil {
		t.Fatalf("dir should be gone")
	}

	return dir
}

func TestStepTempDir(t *testing.T) {
	testStepTempDir_impl(t)
}

func TestStepTempDir_notmpdir(t *testing.T) {
	tempenv := "TMPDIR"
	if runtime.GOOS == "windows" {
		tempenv = "TMP"
	}
	// Verify empty TMPDIR maps to current working directory
	oldenv := os.Getenv(tempenv)
	os.Setenv(tempenv, "")
	defer os.Setenv(tempenv, oldenv)

	dir1 := testStepTempDir_impl(t)

	// Now set TMPDIR to current directory
	abspath, err := filepath.Abs(".")
	if err != nil {
		t.Fatalf("could not get current working directory")
	}
	os.Setenv(tempenv, abspath)

	dir2 := testStepTempDir_impl(t)

	if filepath.Dir(dir1) != filepath.Dir(dir2) {
		t.Fatalf("temp base directories do not match: %s %s", filepath.Dir(dir1), filepath.Dir(dir2))
	}
}

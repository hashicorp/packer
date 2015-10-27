package docker

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
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
	tempenv := "PACKER_TMP_DIR"

	oldenv := os.Getenv(tempenv)
	defer os.Setenv(tempenv, oldenv)
	os.Setenv(tempenv, "")

	dir1 := testStepTempDir_impl(t)

	cd, err := packer.ConfigDir()
	if err != nil {
		t.Fatalf("bad ConfigDir")
	}
	td := filepath.Join(cd, "tmp")
	os.Setenv(tempenv, td)

	dir2 := testStepTempDir_impl(t)

	if filepath.Dir(dir1) != filepath.Dir(dir2) {
		t.Fatalf("temp base directories do not match: %s %s", filepath.Dir(dir1), filepath.Dir(dir2))
	}
}

func TestStepTempDir_packertmpdir(t *testing.T) {
	tempenv := "PACKER_TMP_DIR"

	oldenv := os.Getenv(tempenv)
	defer os.Setenv(tempenv, oldenv)
	os.Setenv(tempenv, ".")

	dir1 := testStepTempDir_impl(t)

	abspath, err := filepath.Abs(".")
	if err != nil {
		t.Fatalf("bad absolute path")
	}
	dir2 := filepath.Join(abspath, "tmp")

	if filepath.Dir(dir1) != filepath.Dir(dir2) {
		t.Fatalf("temp base directories do not match: %s %s", filepath.Dir(dir1), filepath.Dir(dir2))
	}
}

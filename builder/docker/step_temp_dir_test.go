package docker

import (
	"github.com/mitchellh/multistep"
	"os"
	"testing"
)

func TestStepTempDir_impl(t *testing.T) {
	var _ multistep.Step = new(StepTempDir)
}

func TestStepTempDir(t *testing.T) {
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
}

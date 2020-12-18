package docker

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
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
	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}

	// Verify that we got the temp dir
	dirRaw, ok := state.GetOk("temp_dir")
	if !ok {
		t.Fatalf("should've made temp_dir")
	}
	dir := dirRaw.(string)

	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("Stat for %s failed: err: %s", err, dir)
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

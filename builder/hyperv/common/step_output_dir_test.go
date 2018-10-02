package common

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/packer/helper/multistep"
)

func TestStepOutputDir_imp(t *testing.T) {
	var _ multistep.Step = new(StepOutputDir)
}

func TestStepOutputDir_Default(t *testing.T) {
	state := testState(t)
	step := new(StepOutputDir)

	step.Path = genTestDirPath("packerHypervOutput")

	// Test the run
	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("Bad action: %v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("Should NOT have error")
	}

	// The directory should have been created
	if _, err := os.Stat(step.Path); err != nil {
		t.Fatal("Should have created output directory")
	}

	// Remove the directory created due to test
	err := os.RemoveAll(step.Path)
	if err != nil {
		t.Fatalf("Error encountered removing directory created by test: %s", err)
	}
}

func TestStepOutputDir_DirectoryAlreadyExistsNoForce(t *testing.T) {
	state := testState(t)
	step := new(StepOutputDir)

	step.Path = genTestDirPath("packerHypervOutput")

	// Create the directory so that we can test
	err := os.Mkdir(step.Path, 0755)
	if err != nil {
		t.Fatal("Test failed to create directory for test of Cancel and Cleanup")
	}
	defer os.RemoveAll(step.Path) // Ensure we clean up if something goes wrong

	step.Force = false // Default
	// Test the run
	if action := step.Run(context.Background(), state); action != multistep.ActionHalt {
		t.Fatalf("Should halt when directory exists and 'Force' is false. Bad action: %v", action)
	}
	if _, ok := state.GetOk("error"); !ok {
		t.Fatal("Should error when directory exists and 'Force' is false")
	}
}

func TestStepOutputDir_DirectoryAlreadyExistsForce(t *testing.T) {
	state := testState(t)
	step := new(StepOutputDir)

	step.Path = genTestDirPath("packerHypervOutput")

	// Create the directory so that we can test
	err := os.Mkdir(step.Path, 0755)
	if err != nil {
		t.Fatal("Test failed to create directory for test of Cancel and Cleanup")
	}
	defer os.RemoveAll(step.Path) // Ensure we clean up if something goes wrong

	step.Force = true // User specified that existing directory and contents should be discarded
	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("Should complete when directory exists and 'Force' is true. Bad action: %v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatalf("Should NOT error when directory exists and 'Force' is true: %s", err)
	}
}

func TestStepOutputDir_CleanupBuildCancelled(t *testing.T) {
	state := testState(t)
	step := new(StepOutputDir)

	step.Path = genTestDirPath("packerHypervOutput")

	// Create the directory so that we can test the cleanup
	err := os.Mkdir(step.Path, 0755)
	if err != nil {
		t.Fatal("Test failed to create directory for test of Cancel and Cleanup")
	}
	defer os.RemoveAll(step.Path) // Ensure we clean up if something goes wrong

	// 'Cancel' the build
	state.Put(multistep.StateCancelled, true)

	// Ensure the directory isn't removed if the cleanup flag is false
	step.cleanup = false
	step.Cleanup(state)
	if _, err := os.Stat(step.Path); err != nil {
		t.Fatal("Output dir should NOT be removed if on 'Cancel' if cleanup flag is unset/false")
	}

	// Ensure the directory is removed if the cleanup flag is true
	step.cleanup = true
	step.Cleanup(state)
	if _, err := os.Stat(step.Path); err == nil {
		t.Fatalf("Output directory should NOT exist after 'Cancel' and Cleanup: %s", step.Path)
	}
}

func TestStepOutputDir_CleanupBuildHalted(t *testing.T) {
	state := testState(t)
	step := new(StepOutputDir)

	step.Path = genTestDirPath("packerHypervOutput")

	// Create the directory so that we can test the cleanup
	err := os.Mkdir(step.Path, 0755)
	if err != nil {
		t.Fatal("Test failed to create directory for test of Cancel and Cleanup")
	}
	defer os.RemoveAll(step.Path) // Ensure we clean up if something goes wrong

	// 'Halt' the build and test the directory is removed
	state.Put(multistep.StateHalted, true)

	// Ensure the directory isn't removed if the cleanup flag is false
	step.cleanup = false
	step.Cleanup(state)
	if _, err := os.Stat(step.Path); err != nil {
		t.Fatal("Output dir should NOT be removed if on 'Halt' if cleanup flag is unset/false")
	}

	// Ensure the directory is removed if the cleanup flag is true
	step.cleanup = true
	step.Cleanup(state)
	if _, err := os.Stat(step.Path); err == nil {
		t.Fatalf("Output directory should NOT exist after 'Halt' and Cleanup: %s", step.Path)
	}
}

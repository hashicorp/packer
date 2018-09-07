package common

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/hashicorp/packer/helper/multistep"
)

func TestStepCreateBuildDir_imp(t *testing.T) {
	var _ multistep.Step = new(StepCreateBuildDir)
}

func TestStepCreateBuildDir_Defaults(t *testing.T) {
	state := testState(t)
	step := new(StepCreateBuildDir)

	// Default is for the user not to supply value for TempPath. When
	// nothing is set the step should use the OS temp directory as the root
	// for the build directory
	step.TempPath = ""

	// Test the run
	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("Bad action: %v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("Should NOT have error")
	}

	if v, ok := state.GetOk("build_dir"); !ok {
		t.Fatal("Should store path to build directory in statebag as 'build_dir'")
	} else {
		// On windows convert everything to forward slash separated paths
		// This prevents the regexp interpreting backslashes as escape sequences
		stateBuildDir := filepath.ToSlash(v.(string))
		expectedBuildDirRe := regexp.MustCompile(
			filepath.ToSlash(filepath.Join(os.TempDir(), "packerhv") + `[[:digit:]]{9}$`))
		match := expectedBuildDirRe.MatchString(stateBuildDir)
		if !match {
			t.Fatalf("Got path that doesn't match expected format in 'build_dir': %s", stateBuildDir)
		}
	}

	// Test Cleanup
	step.Cleanup(state)
	if _, err := os.Stat(step.buildDir); err == nil {
		t.Fatalf("Build directory should NOT exist after Cleanup: %s", step.buildDir)
	}
}

func TestStepCreateBuildDir_UserDefinedTempPath(t *testing.T) {
	state := testState(t)
	step := new(StepCreateBuildDir)

	// Create a directory we'll use as the user supplied temp_path
	step.TempPath = genTestDirPath("userTempDir")
	err := os.Mkdir(step.TempPath, 0755) // The directory must exist
	if err != nil {
		t.Fatal("Error creating test directory")
	}
	defer os.RemoveAll(step.TempPath)

	// Test the run
	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("Bad action: %v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("Should NOT have error")
	}

	if v, ok := state.GetOk("build_dir"); !ok {
		t.Fatal("Should store path to build directory in statebag as 'build_dir'")
	} else {
		// On windows convert everything to forward slash separated paths
		// This prevents the regexp interpreting backslashes as escape sequences
		stateBuildDir := filepath.ToSlash(v.(string))
		expectedBuildDirRe := regexp.MustCompile(
			filepath.ToSlash(filepath.Join(step.TempPath, "packerhv") + `[[:digit:]]{9}$`))
		match := expectedBuildDirRe.MatchString(stateBuildDir)
		if !match {
			t.Fatalf("Got path that doesn't match expected format in 'build_dir': %s", stateBuildDir)
		}
	}

	// Test Cleanup
	step.Cleanup(state)
	if _, err := os.Stat(step.buildDir); err == nil {
		t.Fatalf("Build directory should NOT exist after Cleanup: %s", step.buildDir)
	}
	if _, err := os.Stat(step.TempPath); err != nil {
		t.Fatal("User supplied root for build directory should NOT be deleted by Cleanup")
	}
}

func TestStepCreateBuildDir_BadTempPath(t *testing.T) {
	state := testState(t)
	step := new(StepCreateBuildDir)

	// Bad
	step.TempPath = genTestDirPath("iDontExist")

	// Test the run
	if action := step.Run(context.Background(), state); action != multistep.ActionHalt {
		t.Fatalf("Bad action: %v", action)
	}
	if _, ok := state.GetOk("error"); !ok {
		t.Fatal("Should have error due to bad path")
	}
}

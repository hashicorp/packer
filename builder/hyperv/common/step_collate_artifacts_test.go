package common

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/hashicorp/packer/helper/multistep"
)

func TestStepCollateArtifacts_impl(t *testing.T) {
	var _ multistep.Step = new(StepCollateArtifacts)
}

func TestStepCollateArtifacts_exportedArtifacts(t *testing.T) {
	state := testState(t)
	step := new(StepCollateArtifacts)

	step.OutputDir = "foopath"
	vmName := "foo"

	// Uses export path from the state bag
	exportPath := filepath.Join(step.OutputDir, vmName)
	state.Put("export_path", exportPath)

	driver := state.Get("driver").(*DriverMock)

	// Test the run
	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("Bad action: %v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("Should NOT have error")
	}

	// Test the driver
	if !driver.PreserveLegacyExportBehaviour_Called {
		t.Fatal("Should have called PreserveLegacyExportBehaviour")
	}
	if driver.PreserveLegacyExportBehaviour_SrcPath != exportPath {
		t.Fatalf("Should call with correct srcPath. Got: %s Wanted: %s",
			driver.PreserveLegacyExportBehaviour_SrcPath, exportPath)
	}
	if driver.PreserveLegacyExportBehaviour_DstPath != step.OutputDir {
		t.Fatalf("Should call with correct dstPath. Got: %s Wanted: %s",
			driver.PreserveLegacyExportBehaviour_DstPath, step.OutputDir)
	}

	// TODO: Create MoveCreatedVHDsToOutput func etc
	// if driver.MoveCreatedVHDsToOutput_Called {
	// t.Fatal("Should NOT have called MoveCreatedVHDsToOutput")
	// }
}

func TestStepCollateArtifacts_skipExportedArtifacts(t *testing.T) {
	state := testState(t)
	step := new(StepCollateArtifacts)

	// TODO: Needs the path to the main output directory
	// outputDir := "foopath"
	// Export has been skipped
	step.SkipExport = true

	driver := state.Get("driver").(*DriverMock)

	// Test the run
	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("Bad action: %v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("Should NOT have error")
	}

	// TODO: Create MoveCreatedVHDsToOutput func etc
	// if !driver.MoveCreatedVHDsToOutput_Called {
	// t.Fatal("Should have called MoveCreatedVHDsToOutput")
	// }

	if driver.PreserveLegacyExportBehaviour_Called {
		t.Fatal("Should NOT have called PreserveLegacyExportBehaviour")
	}
}

package common

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/hashicorp/packer/helper/multistep"
)

func TestStepExportVm_impl(t *testing.T) {
	var _ multistep.Step = new(StepExportVm)
}

func TestStepExportVm(t *testing.T) {
	state := testState(t)
	step := new(StepExportVm)

	// ExportVirtualMachine needs the VM name and a path to export to
	vmName := "foo"
	state.Put("vmName", vmName)
	outputDir := "foopath"
	step.OutputDir = outputDir

	driver := state.Get("driver").(*DriverMock)

	// Test the run
	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("Bad action: %v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("Should NOT have error")
	}

	// Test the driver
	if !driver.ExportVirtualMachine_Called {
		t.Fatal("Should have called ExportVirtualMachine")
	}
	if driver.ExportVirtualMachine_Path != outputDir {
		t.Fatalf("Should call with correct path. Got: %s Wanted: %s",
			driver.ExportVirtualMachine_Path, outputDir)
	}
	if driver.ExportVirtualMachine_VmName != vmName {
		t.Fatalf("Should call with correct vm name. Got: %s Wanted: %s",
			driver.ExportVirtualMachine_VmName, vmName)
	}

	if !driver.PreserveLegacyExportBehaviour_Called {
		t.Fatal("Should have called PreserveLegacyExportBehaviour")
	}
	exportPath := filepath.Join(outputDir, vmName)
	if driver.PreserveLegacyExportBehaviour_SrcPath != exportPath {
		t.Fatalf("Should call with correct srcPath. Got: %s Wanted: %s",
			driver.PreserveLegacyExportBehaviour_SrcPath, exportPath)
	}
	if driver.PreserveLegacyExportBehaviour_DstPath != outputDir {
		t.Fatalf("Should call with correct dstPath. Got: %s Wanted: %s",
			driver.PreserveLegacyExportBehaviour_DstPath, outputDir)
	}

}

func TestStepExportVm_skip(t *testing.T) {
	state := testState(t)
	step := new(StepExportVm)
	step.SkipExport = true

	// ExportVirtualMachine needs the VM name and a path to export to
	vmName := "foo"
	state.Put("vmName", vmName)
	outputDir := "foopath"
	step.OutputDir = outputDir

	driver := state.Get("driver").(*DriverMock)

	// Test the run
	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("Bad action: %v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatalf("Should NOT have error")
	}

	// Test the driver
	if driver.ExportVirtualMachine_Called {
		t.Fatal("Should not have called ExportVirtualMachine")
	}

	if driver.PreserveLegacyExportBehaviour_Called {
		t.Fatal("Should NOT have called PreserveLegacyExportBehaviour")
	}
}

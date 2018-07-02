package common

import (
	"context"
	"testing"

	"github.com/hashicorp/packer/helper/multistep"
)

func TestStepExportVm_impl(t *testing.T) {
	var _ multistep.Step = new(StepCompactDisk)
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
}

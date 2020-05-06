package common

import (
	"context"
	"testing"

	"github.com/hashicorp/packer/helper/multistep"
)

func TestStepExport_impl(t *testing.T) {
	var _ multistep.Step = new(StepExport)
}

func TestStepExport(t *testing.T) {
	state := testState(t)
	step := new(StepExport)

	state.Put("vmName", "foo")

	driver := state.Get("driver").(*DriverMock)

	// Test the run
	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	// Test output state
	if _, ok := state.GetOk("exportPath"); !ok {
		t.Fatal("should set exportPath")
	}

	// Test driver
	if len(driver.VBoxManageCalls) != 2 {
		t.Fatal("should call vboxmanage")
	}
	if driver.VBoxManageCalls[0][0] != "modifyvm" {
		t.Fatal("bad")
	}
	if driver.VBoxManageCalls[1][0] != "export" {
		t.Fatal("bad")
	}
}

func TestStepExport_OutputPath(t *testing.T) {
	type testCase struct {
		Step     *StepExport
		Expected string
		Reason   string
	}
	tcs := []testCase{
		{
			Step: &StepExport{
				Format:         "ova",
				OutputDir:      "output-dir",
				OutputFilename: "output-filename",
			},
			Expected: "output-dir/output-filename.ova",
			Reason:   "output_filename should not be vmName if set.",
		},
		{
			Step: &StepExport{
				Format:         "ovf",
				OutputDir:      "output-dir",
				OutputFilename: "",
			},
			Expected: "output-dir/foo.ovf",
			Reason:   "output_filename should default to vmName.",
		},
	}
	for _, tc := range tcs {
		state := testState(t)
		state.Put("vmName", "foo")

		// Test the run
		if action := tc.Step.Run(context.Background(), state); action != multistep.ActionContinue {
			t.Fatalf("bad action: %#v", action)
		}

		// Test output state
		path, ok := state.GetOk("exportPath")
		if !ok {
			t.Fatal("should set exportPath")
		}
		if path != tc.Expected {
			t.Fatalf("Expected %s didn't match received %s: %s", tc.Expected, path, tc.Reason)
		}
	}
}

func TestStepExport_SkipExport(t *testing.T) {
	state := testState(t)
	step := StepExport{SkipExport: true}

	state.Put("vmName", "foo")

	driver := state.Get("driver").(*DriverMock)

	// Test the run
	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}
	// Test driver
	if len(driver.VBoxManageCalls) != 0 {
		t.Fatal("shouldn't have called vboxmanage; skip_export was set.")
	}

}

package common

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
)

func testOutputDir(t *testing.T) string {
	td, err := ioutil.TempDir("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	os.RemoveAll(td)

	return td
}

func TestStepOutputDir_impl(t *testing.T) {
	var _ multistep.Step = new(StepOutputDir)
}

func TestStepOutputDir(t *testing.T) {
	state := testState(t)
	driver := new(DriverMock)
	state.Put("driver", driver)

	td := testOutputDir(t)
	outconfig := &OutputConfig{
		OutputDir: td,
	}

	step := &StepOutputDir{
		OutputConfig: outconfig,
		VMName:       "testVM",
	}
	// Delete the test output directory when done
	defer os.RemoveAll(td)

	// Test the run
	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}
	if _, err := os.Stat(td); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Test the cleanup
	step.Cleanup(state)
	if _, err := os.Stat(td); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestStepOutputDir_existsNoForce(t *testing.T) {
	state := testState(t)

	td := testOutputDir(t)
	outconfig := &OutputConfig{
		OutputDir: td,
	}

	step := &StepOutputDir{
		OutputConfig: outconfig,
		VMName:       "testVM",
	}
	// Delete the test output directory when done
	defer os.RemoveAll(td)

	// Make sure the dir exists
	if err := os.MkdirAll(td, 0755); err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.RemoveAll(td)

	// Test the run
	if action := step.Run(context.Background(), state); action != multistep.ActionHalt {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); !ok {
		t.Fatal("should have error")
	}

	// Test the cleanup
	step.Cleanup(state)
	if _, err := os.Stat(td); err != nil {
		t.Fatal("should not delete dir")
	}
}

func TestStepOutputDir_existsForce(t *testing.T) {
	state := testState(t)

	td := testOutputDir(t)
	outconfig := &OutputConfig{
		OutputDir: td,
	}

	step := &StepOutputDir{
		OutputConfig: outconfig,
		VMName:       "testVM",
	}
	step.Force = true

	// Delete the test output directory when done
	defer os.RemoveAll(td)

	// Make sure the dir exists
	if err := os.MkdirAll(td, 0755); err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.RemoveAll(td)

	// Test the run
	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}
	if _, err := os.Stat(td); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestStepOutputDir_cancel(t *testing.T) {
	state := testState(t)
	td := testOutputDir(t)
	outconfig := &OutputConfig{
		OutputDir: td,
	}

	step := &StepOutputDir{
		OutputConfig: outconfig,
		VMName:       "testVM",
	}

	// Test the run
	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}
	if _, err := os.Stat(td); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Test cancel/halt
	state.Put(multistep.StateCancelled, true)
	step.Cleanup(state)
	if _, err := os.Stat(td); err == nil {
		t.Fatal("directory should not exist")
	}
}

func TestStepOutputDir_halt(t *testing.T) {
	state := testState(t)
	td := testOutputDir(t)
	outconfig := &OutputConfig{
		OutputDir: td,
	}

	step := &StepOutputDir{
		OutputConfig: outconfig,
		VMName:       "testVM",
	}

	// Test the run
	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}
	if _, err := os.Stat(td); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Test cancel/halt
	state.Put(multistep.StateHalted, true)
	step.Cleanup(state)
	if _, err := os.Stat(td); err == nil {
		t.Fatal("directory should not exist")
	}
}

func TestStepOutputDir_Remote(t *testing.T) {
	// Tests remote driver
	state := testState(t)
	driver := new(RemoteDriverMock)
	state.Put("driver", driver)

	td := testOutputDir(t)
	outconfig := &OutputConfig{
		OutputDir:       td,
		RemoteOutputDir: "remote_path",
	}

	step := &StepOutputDir{
		OutputConfig: outconfig,
		VMName:       "testVM",
		RemoteType:   "esx5",
	}
	// Delete the test output directory when done
	defer os.RemoveAll(td)

	// Test the run
	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}

	// We don't pre-create the output path for export but we do set it in state.
	exportOutputPath := state.Get("export_output_path").(string)
	if exportOutputPath != td {
		t.Fatalf("err: should have set export_output_path!")
	}
}

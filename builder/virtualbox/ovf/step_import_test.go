package ovf

import (
	"github.com/mitchellh/multistep"
	vboxcommon "github.com/mitchellh/packer/builder/virtualbox/common"
	"testing"
)

func TestStepImport_impl(t *testing.T) {
	var _ multistep.Step = new(StepImport)
}

func TestStepImport(t *testing.T) {
	state := testState(t)
	step := new(StepImport)
	step.Name = "bar"
	step.SourcePath = "foo"

	driver := state.Get("driver").(*vboxcommon.DriverMock)

	// Test the run
	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	// Test driver
	if !driver.ImportCalled {
		t.Fatal("import should be called")
	}
	if driver.ImportName != step.Name {
		t.Fatalf("bad: %#v", driver.ImportName)
	}
	if driver.ImportPath != step.SourcePath {
		t.Fatalf("bad: %#v", driver.ImportPath)
	}

	// Test output state
	if name, ok := state.GetOk("vmName"); !ok {
		t.Fatal("vmName should be set")
	} else if name != "bar" {
		t.Fatalf("bad: %#v", name)
	}

	// Test cleanup
	step.Cleanup(state)
	if !driver.DeleteCalled {
		t.Fatal("delete should be called")
	}
	if driver.DeleteName != "bar" {
		t.Fatalf("bad: %#v", driver.DeleteName)
	}
}

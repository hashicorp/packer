package vmx

import (
	"testing"

	"github.com/mitchellh/multistep"
	vmwcommon "github.com/mitchellh/packer/builder/vmware/common"
)

func TestStepCloneVMX_impl(t *testing.T) {
	var _ multistep.Step = new(StepCloneVMX)
}

func TestStepCloneVMX(t *testing.T) {
	state := testState(t)
	step := new(StepCloneVMX)
	step.OutputDir = "/foo"
	step.Path = "/bar/bar.vmx"
	step.VMName = "foo"

	driver := state.Get("driver").(*vmwcommon.DriverMock)

	// Test the run
	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	if !driver.CloneCalled {
		t.Fatal("clone should be called")
	}
}

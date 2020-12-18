package common

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
)

func TestStepCreateVM_impl(t *testing.T) {
	var _ multistep.Step = new(StepCreateVM)
}

func TestStepCreateVM(t *testing.T) {
	state := testState(t)
	step := new(StepCreateVM)

	step.VMName = "test-VM-Name"
	driver := state.Get("driver").(*DriverMock)

	// Test the run
	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("Bad action: %v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("Should NOT have error")
	}

	// Test the driver
	if !driver.CheckVMName_Called {
		t.Fatal("Should have called CheckVMName")
	}
}

func TestStepCreateVM_CheckVMNameErr(t *testing.T) {
	state := testState(t)
	step := new(StepCreateVM)

	step.VMName = "test-VM-Name"
	driver := state.Get("driver").(*DriverMock)
	driver.CheckVMName_Err = fmt.Errorf("A virtual machine with the name is already" +
		" defined in Hyper-V. To avoid a name collision, please set your " +
		"vm_name to a unique value")

	// Test the run
	if action := step.Run(context.Background(), state); action != multistep.ActionHalt {
		t.Fatalf("Bad action: %v", action)
	}
	if _, ok := state.GetOk("error"); !ok {
		t.Fatal("Should have error")
	}

	// Test the driver
	if !driver.CheckVMName_Called {
		t.Fatal("Should have called CheckVMName")
	}
}

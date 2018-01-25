package triton

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/packer/helper/multistep"
)

func TestStepDeleteMachine(t *testing.T) {
	state := testState(t)
	step := new(StepDeleteMachine)
	defer step.Cleanup(state)

	driver := state.Get("driver").(*DriverMock)

	machineId := "test-machine-id"
	state.Put("machine", machineId)

	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}

	step.Cleanup(state)

	if driver.DeleteMachineId != machineId {
		t.Fatalf("should've deleted machine (%s != %s)", driver.DeleteMachineId, machineId)
	}
}

func TestStepDeleteMachine_DeleteMachineError(t *testing.T) {
	state := testState(t)
	step := new(StepDeleteMachine)
	defer step.Cleanup(state)

	driver := state.Get("driver").(*DriverMock)

	machineId := "test-machine-id"
	state.Put("machine", machineId)

	driver.DeleteMachineErr = errors.New("error")

	if action := step.Run(context.Background(), state); action != multistep.ActionHalt {
		t.Fatalf("bad action: %#v", action)
	}

	if _, ok := state.GetOk("error"); !ok {
		t.Fatalf("should have error")
	}

	if _, ok := state.GetOk("machine"); !ok {
		t.Fatalf("should have machine")
	}
}

func TestStepDeleteMachine_WaitForMachineDeletionError(t *testing.T) {
	state := testState(t)
	step := new(StepDeleteMachine)
	defer step.Cleanup(state)

	driver := state.Get("driver").(*DriverMock)

	machineId := "test-machine-id"
	state.Put("machine", machineId)

	driver.WaitForMachineDeletionErr = errors.New("error")

	if action := step.Run(context.Background(), state); action != multistep.ActionHalt {
		t.Fatalf("bad action: %#v", action)
	}

	if _, ok := state.GetOk("error"); !ok {
		t.Fatalf("should have error")
	}

	if _, ok := state.GetOk("machine"); !ok {
		t.Fatalf("should have machine")
	}
}

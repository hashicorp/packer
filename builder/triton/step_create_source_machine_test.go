package triton

import (
	"errors"
	"testing"

	"github.com/mitchellh/multistep"
)

func TestStepCreateSourceMachine(t *testing.T) {
	state := testState(t)
	step := new(StepCreateSourceMachine)
	defer step.Cleanup(state)

	driver := state.Get("driver").(*DriverMock)

	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}

	machineIdRaw, ok := state.GetOk("machine")
	if !ok {
		t.Fatalf("should have machine")
	}

	step.Cleanup(state)

	if driver.DeleteMachineId != machineIdRaw.(string) {
		t.Fatalf("should've deleted machine (%s != %s)", driver.DeleteMachineId, machineIdRaw.(string))
	}
}

func TestStepCreateSourceMachine_CreateMachineError(t *testing.T) {
	state := testState(t)
	step := new(StepCreateSourceMachine)
	defer step.Cleanup(state)

	driver := state.Get("driver").(*DriverMock)

	driver.CreateMachineErr = errors.New("error")

	if action := step.Run(state); action != multistep.ActionHalt {
		t.Fatalf("bad action: %#v", action)
	}

	if _, ok := state.GetOk("error"); !ok {
		t.Fatalf("should have error")
	}

	if _, ok := state.GetOk("machine"); ok {
		t.Fatalf("should NOT have machine")
	}
}

func TestStepCreateSourceMachine_WaitForMachineStateError(t *testing.T) {
	state := testState(t)
	step := new(StepCreateSourceMachine)
	defer step.Cleanup(state)

	driver := state.Get("driver").(*DriverMock)

	driver.WaitForMachineStateErr = errors.New("error")

	if action := step.Run(state); action != multistep.ActionHalt {
		t.Fatalf("bad action: %#v", action)
	}

	if _, ok := state.GetOk("error"); !ok {
		t.Fatalf("should have error")
	}

	if _, ok := state.GetOk("machine"); ok {
		t.Fatalf("should NOT have machine")
	}
}

func TestStepCreateSourceMachine_StopMachineError(t *testing.T) {
	state := testState(t)
	step := new(StepCreateSourceMachine)
	defer step.Cleanup(state)

	driver := state.Get("driver").(*DriverMock)

	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}

	_, ok := state.GetOk("machine")
	if !ok {
		t.Fatalf("should have machine")
	}

	driver.StopMachineErr = errors.New("error")
	step.Cleanup(state)

	if _, ok := state.GetOk("error"); !ok {
		t.Fatalf("should have error")
	}

	if _, ok := state.GetOk("machine"); !ok {
		t.Fatalf("should have machine")
	}
}

func TestStepCreateSourceMachine_WaitForMachineStoppedError(t *testing.T) {
	state := testState(t)
	step := new(StepCreateSourceMachine)
	defer step.Cleanup(state)

	driver := state.Get("driver").(*DriverMock)

	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}

	_, ok := state.GetOk("machine")
	if !ok {
		t.Fatalf("should have machine")
	}

	driver.WaitForMachineStateErr = errors.New("error")
	step.Cleanup(state)

	if _, ok := state.GetOk("error"); !ok {
		t.Fatalf("should have error")
	}

	if _, ok := state.GetOk("machine"); !ok {
		t.Fatalf("should have machine")
	}
}

func TestStepCreateSourceMachine_DeleteMachineError(t *testing.T) {
	state := testState(t)
	step := new(StepCreateSourceMachine)
	defer step.Cleanup(state)

	driver := state.Get("driver").(*DriverMock)

	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}

	_, ok := state.GetOk("machine")
	if !ok {
		t.Fatalf("should have machine")
	}

	driver.DeleteMachineErr = errors.New("error")
	step.Cleanup(state)

	if _, ok := state.GetOk("error"); !ok {
		t.Fatalf("should have error")
	}

	if _, ok := state.GetOk("machine"); !ok {
		t.Fatalf("should have machine")
	}
}

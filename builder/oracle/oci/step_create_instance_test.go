package oci

import (
	"errors"
	"testing"

	"github.com/hashicorp/packer/helper/multistep"
)

func TestStepCreateInstance(t *testing.T) {
	state := testState()
	state.Put("publicKey", "key")

	step := new(stepCreateInstance)
	defer step.Cleanup(state)

	driver := state.Get("driver").(*driverMock)

	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}

	instanceIDRaw, ok := state.GetOk("instance_id")
	if !ok {
		t.Fatalf("should have machine")
	}

	step.Cleanup(state)

	if driver.TerminateInstanceID != instanceIDRaw.(string) {
		t.Fatalf(
			"should've deleted instance (%s != %s)",
			driver.TerminateInstanceID, instanceIDRaw.(string))
	}
}

func TestStepCreateInstance_CreateInstanceErr(t *testing.T) {
	state := testState()
	state.Put("publicKey", "key")

	step := new(stepCreateInstance)
	defer step.Cleanup(state)

	driver := state.Get("driver").(*driverMock)
	driver.CreateInstanceErr = errors.New("error")

	if action := step.Run(state); action != multistep.ActionHalt {
		t.Fatalf("bad action: %#v", action)
	}

	if _, ok := state.GetOk("error"); !ok {
		t.Fatalf("should have error")
	}

	if _, ok := state.GetOk("instance_id"); ok {
		t.Fatalf("should NOT have instance_id")
	}

	step.Cleanup(state)

	if driver.TerminateInstanceID != "" {
		t.Fatalf("Should not have tried to terminate an instance")
	}
}

func TestStepCreateInstance_WaitForInstanceStateErr(t *testing.T) {
	state := testState()
	state.Put("publicKey", "key")

	step := new(stepCreateInstance)
	defer step.Cleanup(state)

	driver := state.Get("driver").(*driverMock)
	driver.WaitForInstanceStateErr = errors.New("error")

	if action := step.Run(state); action != multistep.ActionHalt {
		t.Fatalf("bad action: %#v", action)
	}

	if _, ok := state.GetOk("error"); !ok {
		t.Fatalf("should have error")
	}
}

func TestStepCreateInstance_TerminateInstanceErr(t *testing.T) {
	state := testState()
	state.Put("publicKey", "key")

	step := new(stepCreateInstance)
	defer step.Cleanup(state)

	driver := state.Get("driver").(*driverMock)

	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}

	_, ok := state.GetOk("instance_id")
	if !ok {
		t.Fatalf("should have machine")
	}

	driver.TerminateInstanceErr = errors.New("error")
	step.Cleanup(state)

	if _, ok := state.GetOk("error"); !ok {
		t.Fatalf("should have error")
	}
}

func TestStepCreateInstanceCleanup_WaitForInstanceStateErr(t *testing.T) {
	state := testState()
	state.Put("publicKey", "key")

	step := new(stepCreateInstance)
	defer step.Cleanup(state)

	driver := state.Get("driver").(*driverMock)

	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}

	driver.WaitForInstanceStateErr = errors.New("error")
	step.Cleanup(state)

	if _, ok := state.GetOk("error"); !ok {
		t.Fatalf("should have error")
	}
}

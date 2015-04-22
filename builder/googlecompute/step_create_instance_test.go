package googlecompute

import (
	"errors"
	"github.com/mitchellh/multistep"
	"testing"
	"time"
)

func TestStepCreateInstance_impl(t *testing.T) {
	var _ multistep.Step = new(StepCreateInstance)
}

func TestStepCreateInstance(t *testing.T) {
	state := testState(t)
	step := new(StepCreateInstance)
	defer step.Cleanup(state)

	state.Put("ssh_public_key", "key")

	config := state.Get("config").(*Config)
	driver := state.Get("driver").(*DriverMock)

	// run the step
	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}

	// Verify state
	nameRaw, ok := state.GetOk("instance_name")
	if !ok {
		t.Fatal("should have instance name")
	}

	// cleanup
	step.Cleanup(state)

	if driver.DeleteInstanceName != nameRaw.(string) {
		t.Fatal("should've deleted instance")
	}
	if driver.DeleteInstanceZone != config.Zone {
		t.Fatalf("bad instance zone: %#v", driver.DeleteInstanceZone)
	}

	if driver.DeleteDiskName != config.InstanceName {
		t.Fatal("should've deleted disk")
	}
	if driver.DeleteDiskZone != config.Zone {
		t.Fatalf("bad disk zone: %#v", driver.DeleteDiskZone)
	}
}

func TestStepCreateInstance_error(t *testing.T) {
	state := testState(t)
	step := new(StepCreateInstance)
	defer step.Cleanup(state)

	state.Put("ssh_public_key", "key")

	driver := state.Get("driver").(*DriverMock)
	driver.RunInstanceErr = errors.New("error")

	// run the step
	if action := step.Run(state); action != multistep.ActionHalt {
		t.Fatalf("bad action: %#v", action)
	}

	// Verify state
	if _, ok := state.GetOk("error"); !ok {
		t.Fatal("should have error")
	}
	if _, ok := state.GetOk("instance_name"); ok {
		t.Fatal("should NOT have instance name")
	}
}

func TestStepCreateInstance_errorOnChannel(t *testing.T) {
	state := testState(t)
	step := new(StepCreateInstance)
	defer step.Cleanup(state)

	errCh := make(chan error, 1)
	errCh <- errors.New("error")

	state.Put("ssh_public_key", "key")

	driver := state.Get("driver").(*DriverMock)
	driver.RunInstanceErrCh = errCh

	// run the step
	if action := step.Run(state); action != multistep.ActionHalt {
		t.Fatalf("bad action: %#v", action)
	}

	// Verify state
	if _, ok := state.GetOk("error"); !ok {
		t.Fatal("should have error")
	}
	if _, ok := state.GetOk("instance_name"); ok {
		t.Fatal("should NOT have instance name")
	}
}

func TestStepCreateInstance_errorTimeout(t *testing.T) {
	state := testState(t)
	step := new(StepCreateInstance)
	defer step.Cleanup(state)

	errCh := make(chan error, 1)
	go func() {
		<-time.After(10 * time.Millisecond)
		errCh <- nil
	}()

	state.Put("ssh_public_key", "key")

	config := state.Get("config").(*Config)
	config.stateTimeout = 1 * time.Microsecond

	driver := state.Get("driver").(*DriverMock)
	driver.RunInstanceErrCh = errCh

	// run the step
	if action := step.Run(state); action != multistep.ActionHalt {
		t.Fatalf("bad action: %#v", action)
	}

	// Verify state
	if _, ok := state.GetOk("error"); !ok {
		t.Fatal("should have error")
	}
	if _, ok := state.GetOk("instance_name"); ok {
		t.Fatal("should NOT have instance name")
	}
}

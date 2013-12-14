package googlecompute

import (
	"errors"
	"github.com/mitchellh/multistep"
	"testing"
	"time"
)

func TestStepRegisterImage_impl(t *testing.T) {
	var _ multistep.Step = new(StepRegisterImage)
}

func TestStepRegisterImage(t *testing.T) {
	state := testState(t)
	step := new(StepRegisterImage)
	defer step.Cleanup(state)

	config := state.Get("config").(*Config)
	driver := state.Get("driver").(*DriverMock)

	// run the step
	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}

	// Verify state
	if driver.CreateImageName != config.ImageName {
		t.Fatalf("bad: %#v", driver.CreateImageName)
	}
	if driver.CreateImageDesc != config.ImageDescription {
		t.Fatalf("bad: %#v", driver.CreateImageDesc)
	}

	nameRaw, ok := state.GetOk("image_name")
	if !ok {
		t.Fatal("should have name")
	}
	if name, ok := nameRaw.(string); !ok {
		t.Fatal("name is not a string")
	} else if name != config.ImageName {
		t.Fatalf("bad name: %s", name)
	}
}

func TestStepRegisterImage_waitError(t *testing.T) {
	state := testState(t)
	step := new(StepRegisterImage)
	defer step.Cleanup(state)

	errCh := make(chan error, 1)
	errCh <- errors.New("error")

	driver := state.Get("driver").(*DriverMock)
	driver.CreateImageErrCh = errCh

	// run the step
	if action := step.Run(state); action != multistep.ActionHalt {
		t.Fatalf("bad action: %#v", action)
	}

	// Verify state
	if _, ok := state.GetOk("error"); !ok {
		t.Fatal("should have error")
	}
	if _, ok := state.GetOk("image_name"); ok {
		t.Fatal("should NOT have image_name")
	}
}

func TestStepRegisterImage_errorTimeout(t *testing.T) {
	state := testState(t)
	step := new(StepRegisterImage)
	defer step.Cleanup(state)

	errCh := make(chan error, 1)
	go func() {
		<-time.After(10 * time.Millisecond)
		errCh <- nil
	}()

	config := state.Get("config").(*Config)
	config.stateTimeout = 1 * time.Microsecond

	driver := state.Get("driver").(*DriverMock)
	driver.CreateImageErrCh = errCh

	// run the step
	if action := step.Run(state); action != multistep.ActionHalt {
		t.Fatalf("bad action: %#v", action)
	}

	// Verify state
	if _, ok := state.GetOk("error"); !ok {
		t.Fatal("should have error")
	}
	if _, ok := state.GetOk("image_name"); ok {
		t.Fatal("should NOT have image name")
	}
}

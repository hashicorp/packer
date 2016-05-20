package googlecompute

import (
	"errors"
	"testing"

	"github.com/mitchellh/multistep"
)

func TestStepCreateImage_impl(t *testing.T) {
	var _ multistep.Step = new(StepCreateImage)
}

func TestStepCreateImage(t *testing.T) {
	state := testState(t)
	step := new(StepCreateImage)
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
	if driver.CreateImageFamily != config.ImageFamily {
		t.Fatalf("bad: %#v", driver.CreateImageFamily)
	}
	if driver.CreateImageZone != config.Zone {
		t.Fatalf("bad: %#v", driver.CreateImageZone)
	}
	if driver.CreateImageDisk != config.DiskName {
		t.Fatalf("bad: %#v", driver.CreateImageDisk)
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

func TestStepCreateImage_errorOnChannel(t *testing.T) {
	state := testState(t)
	step := new(StepCreateImage)
	defer step.Cleanup(state)

	errCh := make(chan error, 1)
	errCh <- errors.New("error")

	driver := state.Get("driver").(*DriverMock)
	driver.CreateImageErrCh = errCh

	// run the step
	if action := step.Run(state); action != multistep.ActionHalt {
		t.Fatalf("bad action: %#v", action)
	}

	if _, ok := state.GetOk("error"); !ok {
		t.Fatal("should have error")
	}
	if _, ok := state.GetOk("image_name"); ok {
		t.Fatal("should NOT have image")
	}
}

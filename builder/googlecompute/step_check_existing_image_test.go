package googlecompute

import (
	"github.com/mitchellh/multistep"
	"testing"
)

func TestStepCheckExistingImage_impl(t *testing.T) {
	var _ multistep.Step = new(StepCheckExistingImage)
}

func TestStepCheckExistingImage(t *testing.T) {
	state := testState(t)
	step := new(StepCheckExistingImage)
	defer step.Cleanup(state)

	state.Put("instance_name", "foo")

	config := state.Get("config").(*Config)
	driver := state.Get("driver").(*DriverMock)
	driver.ImageExistsResult = true

	// run the step
	if action := step.Run(state); action != multistep.ActionHalt {
		t.Fatalf("bad action: %#v", action)
	}

	// Verify state
	if driver.ImageExistsName != config.ImageName {
		t.Fatalf("bad: %#v", driver.ImageExistsName)
	}
}

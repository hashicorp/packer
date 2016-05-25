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
	driver.CreateImageProjectId = "createimage-project"
	driver.CreateImageSizeGb = 100

	// run the step
	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	
	uncastImage, ok := state.GetOk("image")
	if !ok {
		t.Fatal("should have image")
	}
	image, ok := uncastImage.(Image)
	if !ok {
		t.Fatal("image is not an Image")
	}
	
	// Verify created Image results.
	if image.Name != config.ImageName {
		t.Fatalf("Created image name, %s, does not match config name, %s.", image.Name, config.ImageName)
	}
	if driver.CreateImageProjectId != image.ProjectId {
		t.Fatalf("Created image project ID, %s, does not match driver project ID, %s.", image.ProjectId, driver.CreateImageProjectId)
	}
	if driver.CreateImageSizeGb != image.SizeGb {
		t.Fatalf("Created image size, %d, does not match the expected test value, %d.", image.SizeGb, driver.CreateImageSizeGb)
	}

	// Verify proper args passed to driver.CreateImage.
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

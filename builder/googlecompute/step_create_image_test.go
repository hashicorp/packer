package googlecompute

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/stretchr/testify/assert"
)

func TestStepCreateImage_impl(t *testing.T) {
	var _ multistep.Step = new(StepCreateImage)
}

func TestStepCreateImage(t *testing.T) {
	state := testState(t)
	step := new(StepCreateImage)
	defer step.Cleanup(state)

	c := state.Get("config").(*Config)
	d := state.Get("driver").(*DriverMock)

	// These are the values of the image the driver will return.
	d.CreateImageResultProjectId = "test-project"
	d.CreateImageResultSizeGb = 100

	// run the step
	action := step.Run(context.Background(), state)
	assert.Equal(t, action, multistep.ActionContinue, "Step did not pass.")

	uncastImage, ok := state.GetOk("image")
	assert.True(t, ok, "State does not have resulting image.")
	image, ok := uncastImage.(*Image)
	assert.True(t, ok, "Image in state is not an Image.")

	// Verify created Image results.
	assert.Equal(t, image.Name, c.ImageName, "Created image does not match config name.")
	assert.Equal(t, image.ProjectId, d.CreateImageResultProjectId, "Created image project does not match driver project.")
	assert.Equal(t, image.SizeGb, d.CreateImageResultSizeGb, "Created image size does not match the size returned by the driver.")

	// Verify proper args passed to driver.CreateImage.
	assert.Equal(t, d.CreateImageName, c.ImageName, "Incorrect image name passed to driver.")
	assert.Equal(t, d.CreateImageDesc, c.ImageDescription, "Incorrect image description passed to driver.")
	assert.Equal(t, d.CreateImageFamily, c.ImageFamily, "Incorrect image family passed to driver.")
	assert.Equal(t, d.CreateImageZone, c.Zone, "Incorrect image zone passed to driver.")
	assert.Equal(t, d.CreateImageDisk, c.DiskName, "Incorrect disk passed to driver.")
	assert.Equal(t, d.CreateImageLabels, c.ImageLabels, "Incorrect image_labels passed to driver.")
	assert.Equal(t, d.CreateImageLicenses, c.ImageLicenses, "Incorrect image_licenses passed to driver.")
	assert.Equal(t, d.CreateImageEncryptionKey, c.ImageEncryptionKey.ComputeType(), "Incorrect image_encryption_key passed to driver.")
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
	action := step.Run(context.Background(), state)
	assert.Equal(t, action, multistep.ActionHalt, "Step should not have passed.")
	_, ok := state.GetOk("error")
	assert.True(t, ok, "State should have an error.")
	_, ok = state.GetOk("image_name")
	assert.False(t, ok, "State should not have a resulting image.")
}

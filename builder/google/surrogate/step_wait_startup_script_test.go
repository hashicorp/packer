package surrogate

import (
	"context"
	"testing"

	"github.com/hashicorp/packer/builder/google/gcp"

	"github.com/stretchr/testify/assert"

	"github.com/hashicorp/packer/helper/multistep"
)

func TestStepWaitStartupScript(t *testing.T) {
	state := testState(t)
	step := new(StepWaitStartupScript)
	c := state.Get("config").(*Config)
	d := state.Get("driver").(*gcp.DriverMock)

	testZone := "test-zone"
	testInstanceName := "test-instance-name"

	c.Zone = testZone
	state.Put("instance_name", testInstanceName)

	// This step stops when it gets Done back from the metadata.
	d.GetInstanceMetadataResult = gcp.StartupScriptStatusDone

	// Run the step.
	assert.Equal(t, step.Run(context.Background(), state), multistep.ActionContinue, "Step should have passed and continued.")

	// Check that GetInstanceMetadata was called properly.
	assert.Equal(t, d.GetInstanceMetadataZone, testZone, "Incorrect zone passed to GetInstanceMetadata.")
	assert.Equal(t, d.GetInstanceMetadataName, testInstanceName, "Incorrect instance name passed to GetInstanceMetadata.")
}

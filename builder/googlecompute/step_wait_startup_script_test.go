package googlecompute

import (
	"context"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/stretchr/testify/assert"
)

func TestStepWaitStartupScript(t *testing.T) {
	state := testState(t)
	step := new(StepWaitStartupScript)
	c := state.Get("config").(*Config)
	d := state.Get("driver").(*DriverMock)

	testZone := "test-zone"
	testInstanceName := "test-instance-name"

	c.Zone = testZone
	state.Put("instance_name", testInstanceName)

	// This step stops when it gets Done back from the metadata.
	d.GetInstanceMetadataResult = StartupScriptStatusDone

	// Run the step.
	assert.Equal(t, step.Run(context.Background(), state), multistep.ActionContinue, "Step should have passed and continued.")

	// Check that GetInstanceMetadata was called properly.
	assert.Equal(t, d.GetInstanceMetadataZone, testZone, "Incorrect zone passed to GetInstanceMetadata.")
	assert.Equal(t, d.GetInstanceMetadataName, testInstanceName, "Incorrect instance name passed to GetInstanceMetadata.")
}

func TestStepWaitStartupScript_withWrapStartupScript(t *testing.T) {
	tt := []struct {
		WrapStartup                config.Trilean
		Result, Zone, MetadataName string
	}{
		{WrapStartup: config.TriTrue, Result: StartupScriptStatusDone, Zone: "test-zone", MetadataName: "test-instance-name"},
		{WrapStartup: config.TriFalse},
	}

	for _, tc := range tt {
		tc := tc
		state := testState(t)
		step := new(StepWaitStartupScript)
		c := state.Get("config").(*Config)
		d := state.Get("driver").(*DriverMock)

		c.StartupScriptFile = "startup.sh"
		c.WrapStartupScriptFile = tc.WrapStartup
		c.Zone = "test-zone"
		state.Put("instance_name", "test-instance-name")

		// This step stops when it gets Done back from the metadata.
		d.GetInstanceMetadataResult = tc.Result

		// Run the step.
		assert.Equal(t, step.Run(context.Background(), state), multistep.ActionContinue, "Step should have continued.")

		assert.Equal(t, d.GetInstanceMetadataResult, tc.Result, "MetadataResult was not the expected value.")
		assert.Equal(t, d.GetInstanceMetadataZone, tc.Zone, "Zone was not the expected value.")
		assert.Equal(t, d.GetInstanceMetadataName, tc.MetadataName, "Instance name was not the expected value.")
	}
}

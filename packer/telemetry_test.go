// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package packer

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlattenConfigKeys_nil(t *testing.T) {
	f := flattenConfigKeys(nil)
	assert.Zero(t, f, "Expected empty list.")
}

func TestFlattenConfigKeys_nested(t *testing.T) {
	inp := make(map[string]interface{})
	inp["A"] = ""
	inp["B"] = []string{}

	c := make(map[string]interface{})
	c["X"] = ""
	d := make(map[string]interface{})
	d["a"] = ""

	c["Y"] = d
	inp["C"] = c

	assert.Equal(t,
		[]string{"A", "B", "C/X", "C/Y/a"},
		flattenConfigKeys(inp),
		"Input didn't flatten correctly.",
	)
}

func TestCheckpointTelemetry(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Error("a noop CheckpointTelemetry should not to panic but it did\n", r)
		}
	}()

	// A null CheckpointTelemetry obtained in Packer when the CHECKPOINT_DISABLE env var is set results in a NOOP reporter
	// The null reporter can be executable as a configured reporter but does not report any telemetry data.
	var c *CheckpointTelemetry
	c.SetTemplateType(HCL2Template)
	c.SetBundledUsage()
	c.AddSpan("mockprovisioner", "provisioner", nil)
	if err := c.ReportPanic("Bogus Panic"); err != nil {
		t.Errorf("calling ReportPanic on a nil checkpoint reporter should not error")
	}
	if err := c.Finalize("test", 1, errors.New("Bogus Error")); err != nil {
		t.Errorf("calling Finalize on a nil checkpoint reporter should not error")
	}
}

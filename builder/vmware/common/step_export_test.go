package common

import (
	"context"
	"testing"

	"github.com/hashicorp/packer/helper/multistep"
)

func TestStepExport_impl(t *testing.T) {
	var _ multistep.Step = new(StepExport)
}

func testStepExport_wrongtype_impl(t *testing.T, remoteType string) {
	state := testState(t)
	step := new(StepExport)

	var config DriverConfig
	config.RemoteType = "foo"
	state.Put("driverConfig", &config)

	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	// Cleanup
	step.Cleanup(state)
}

func TestStepExport_wrongtype_impl(t *testing.T) {
	testStepExport_wrongtype_impl(t, "foo")
	testStepExport_wrongtype_impl(t, "")
}

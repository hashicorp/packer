package iso

import (
	"github.com/mitchellh/multistep"
	"testing"
)

func TestStepExport_impl(t *testing.T) {
	var _ multistep.Step = new(StepExport)
}

func testStepExport_wrongtype_impl(t *testing.T, remoteType string) {
	state := testState(t)
	step := new(StepExport)

	var config Config
	config.RemoteType = "foo"
	state.Put("config", &config)

	if action := step.Run(state); action != multistep.ActionContinue {
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

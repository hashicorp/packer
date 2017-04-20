package communicator

import (
	"bytes"
	"testing"

	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
)

func TestStepConnect_impl(t *testing.T) {
	var _ multistep.Step = new(StepConnect)
}

func TestStepConnect_none(t *testing.T) {
	state := testState(t)

	step := &StepConnect{
		Config: &Config{
			Type: "none",
		},
	}
	defer step.Cleanup(state)

	// run the step
	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
}

func testState(t *testing.T) multistep.StateBag {
	state := new(multistep.BasicStateBag)
	state.Put("hook", &packer.MockHook{})
	state.Put("ui", &packer.BasicUi{
		Reader: new(bytes.Buffer),
		Writer: new(bytes.Buffer),
	})
	return state
}

package common

import (
	"sort"
	"testing"

	"github.com/mitchellh/multistep"
)

func TestStepConfigureVM_impl(t *testing.T) {
	var _ multistep.Step = new(StepConfigureVM)
}

func TestStepConfigureVM_Change(t *testing.T) {
	state := testState(t)
	driver := new(DriverMock)
	state.Put("driver", driver)

	step := new(StepConfigureVM)
	step.CustomData = map[string]string{"foo": "bar", "foo1": "bar1"}

	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	if !driver.VMChangeCalled {
		t.Fatal("VM change should be called")
	}

	vco := driver.VMChangeOption
	sort.Strings(vco)
	if vco[0] != "foo1=bar1" || vco[1] != "foo=bar" || vco[2] != "msg.autoanswer=true" || len(vco) != 3 {
		t.Fatalf("wrong return from VM change: %#v", vco)
	}

	// Cleanup
	step.Cleanup(state)
}

func TestStepConfigureVM_NoChange(t *testing.T) {
	state := testState(t)
	driver := new(DriverMock)
	state.Put("driver", driver)

	step := new(StepConfigureVM)

	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	if !driver.VMChangeCalled {
		t.Fatal("VM change should be called")
	}

	vco := driver.VMChangeOption
	sort.Strings(vco)
	if vco[0] != "msg.autoanswer=true" || len(vco) != 1 {
		t.Fatalf("wrong return from VM change: %#v", vco)
	}
	// Cleanup
	step.Cleanup(state)
}

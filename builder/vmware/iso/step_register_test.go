package iso

import (
	"github.com/mitchellh/multistep"
	"testing"
)

func TestStepRegister_impl(t *testing.T) {
	var _ multistep.Step = new(StepRegister)
}

func TestStepRegister_regularDriver(t *testing.T) {
	state := testState(t)
	step := new(StepRegister)

	state.Put("vmx_path", "foo")

	// Test the run
	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	// Cleanup
	step.Cleanup(state)
}

func TestStepRegister_remoteDriver(t *testing.T) {
	state := testState(t)
	step := new(StepRegister)

	driver := new(RemoteDriverMock)
	var config Config
	config.KeepRegistered = false
	state.Put("config", &config)

	state.Put("driver", driver)
	state.Put("vmx_path", "foo")

	// Test the run
	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	// verify
	if !driver.RegisterCalled {
		t.Fatal("register should be called")
	}
	if driver.RegisterPath != "foo" {
		t.Fatal("should call with correct path")
	}
	if driver.UnregisterCalled {
		t.Fatal("unregister should not be called")
	}

	// cleanup
	step.Cleanup(state)
	if !driver.UnregisterCalled {
		t.Fatal("unregister should be called")
	}
	if driver.UnregisterPath != "foo" {
		t.Fatal("should unregister proper path")
	}
}
func TestStepRegister_WithoutUnregister_remoteDriver(t *testing.T) {
	state := testState(t)
	step := new(StepRegister)

	driver := new(RemoteDriverMock)
	var config Config
	config.KeepRegistered = true
	state.Put("config", &config)

	state.Put("driver", driver)
	state.Put("vmx_path", "foo")

	// Test the run
	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	// cleanup
	step.Cleanup(state)
	if driver.UnregisterCalled {
		t.Fatal("unregister should not be called")
	}
}

package docker

import (
	"errors"
	"github.com/mitchellh/multistep"
	"testing"
)

func TestStepPull_impl(t *testing.T) {
	var _ multistep.Step = new(StepPull)
}

func TestStepPull(t *testing.T) {
	state := testState(t)
	step := new(StepPull)
	defer step.Cleanup(state)

	config := state.Get("config").(*Config)
	driver := state.Get("driver").(*MockDriver)

	// run the step
	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}

	// verify we did the right thing
	if !driver.PullCalled {
		t.Fatal("should've pulled")
	}
	if driver.PullImage != config.Image {
		t.Fatalf("bad: %#v", driver.PullImage)
	}
}

func TestStepPull_error(t *testing.T) {
	state := testState(t)
	step := new(StepPull)
	defer step.Cleanup(state)

	driver := state.Get("driver").(*MockDriver)
	driver.PullError = errors.New("foo")

	// run the step
	if action := step.Run(state); action != multistep.ActionHalt {
		t.Fatalf("bad action: %#v", action)
	}

	// verify we have an error
	if _, ok := state.GetOk("error"); !ok {
		t.Fatal("should have error")
	}
}

func TestStepPull_login(t *testing.T) {
	state := testState(t)
	step := new(StepPull)
	defer step.Cleanup(state)

	config := state.Get("config").(*Config)
	driver := state.Get("driver").(*MockDriver)

	config.Login = true

	// run the step
	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}

	// verify we pulled
	if !driver.PullCalled {
		t.Fatal("should've pulled")
	}

	// verify we logged in
	if !driver.LoginCalled {
		t.Fatal("should've logged in")
	}
	if !driver.LogoutCalled {
		t.Fatal("should've logged out")
	}
}

func TestStepPull_noPull(t *testing.T) {
	state := testState(t)
	step := new(StepPull)
	defer step.Cleanup(state)

	config := state.Get("config").(*Config)
	config.Pull = false

	driver := state.Get("driver").(*MockDriver)

	// run the step
	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}

	// verify we did the right thing
	if driver.PullCalled {
		t.Fatal("shouldn't have pulled")
	}
}

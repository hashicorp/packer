package common

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/packer/helper/multistep"
)

func TestStepSuppressMessages_impl(t *testing.T) {
	var _ multistep.Step = new(StepSuppressMessages)
}

func TestStepSuppressMessages(t *testing.T) {
	state := testState(t)
	step := new(StepSuppressMessages)

	driver := state.Get("driver").(*DriverMock)

	// Test the run
	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	if !driver.SuppressMessagesCalled {
		t.Fatal("should call suppressmessages")
	}
}

func TestStepSuppressMessages_error(t *testing.T) {
	state := testState(t)
	step := new(StepSuppressMessages)

	driver := state.Get("driver").(*DriverMock)
	driver.SuppressMessagesErr = errors.New("foo")

	// Test the run
	if action := step.Run(context.Background(), state); action != multistep.ActionHalt {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); !ok {
		t.Fatal("should have error")
	}

	if !driver.SuppressMessagesCalled {
		t.Fatal("should call suppressmessages")
	}
}

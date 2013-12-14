package googlecompute

import (
	"strings"
	"testing"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

func TestStepUpdateGsutil_impl(t *testing.T) {
	var _ multistep.Step = new(StepUpdateGsutil)
}

func TestStepUpdateGsutil(t *testing.T) {
	state := testState(t)
	step := new(StepUpdateGsutil)
	defer step.Cleanup(state)

	comm := new(packer.MockCommunicator)
	state.Put("communicator", comm)

	// run the step
	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}

	// Verify
	if !comm.StartCalled {
		t.Fatal("start should be called")
	}
	if strings.HasPrefix(comm.StartCmd.Command, "sudo") {
		t.Fatal("should not sudo")
	}
	if !strings.Contains(comm.StartCmd.Command, "gsutil update") {
		t.Fatalf("bad command: %#v", comm.StartCmd.Command)
	}
}

func TestStepUpdateGsutil_badExitStatus(t *testing.T) {
	state := testState(t)
	step := new(StepUpdateGsutil)
	defer step.Cleanup(state)

	comm := new(packer.MockCommunicator)
	comm.StartExitStatus = 12
	state.Put("communicator", comm)

	// run the step
	if action := step.Run(state); action != multistep.ActionHalt {
		t.Fatalf("bad action: %#v", action)
	}

	if _, ok := state.GetOk("error"); !ok {
		t.Fatal("should have error")
	}
}

func TestStepUpdateGsutil_nonRoot(t *testing.T) {
	state := testState(t)
	step := new(StepUpdateGsutil)
	defer step.Cleanup(state)

	comm := new(packer.MockCommunicator)
	state.Put("communicator", comm)

	config := state.Get("config").(*Config)
	config.SSHUsername = "bob"

	// run the step
	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}

	// Verify
	if !comm.StartCalled {
		t.Fatal("start should be called")
	}
	if !strings.HasPrefix(comm.StartCmd.Command, "sudo") {
		t.Fatal("should sudo")
	}
	if !strings.Contains(comm.StartCmd.Command, "gsutil update") {
		t.Fatalf("bad command: %#v", comm.StartCmd.Command)
	}
}

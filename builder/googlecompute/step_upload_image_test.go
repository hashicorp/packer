package googlecompute

import (
	"strings"
	"testing"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

func TestStepUploadImage_impl(t *testing.T) {
	var _ multistep.Step = new(StepUploadImage)
}

func TestStepUploadImage(t *testing.T) {
	state := testState(t)
	step := new(StepUploadImage)
	defer step.Cleanup(state)

	comm := new(packer.MockCommunicator)
	state.Put("communicator", comm)
	state.Put("image_file_name", "foo")

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
	if !strings.Contains(comm.StartCmd.Command, "gsutil cp") {
		t.Fatalf("bad command: %#v", comm.StartCmd.Command)
	}
}

func TestStepUploadImage_badExitStatus(t *testing.T) {
	state := testState(t)
	step := new(StepUploadImage)
	defer step.Cleanup(state)

	comm := new(packer.MockCommunicator)
	comm.StartExitStatus = 12
	state.Put("communicator", comm)
	state.Put("image_file_name", "foo")

	// run the step
	if action := step.Run(state); action != multistep.ActionHalt {
		t.Fatalf("bad action: %#v", action)
	}

	if _, ok := state.GetOk("error"); !ok {
		t.Fatal("should have error")
	}
}

func TestStepUploadImage_nonRoot(t *testing.T) {
	state := testState(t)
	step := new(StepUploadImage)
	defer step.Cleanup(state)

	comm := new(packer.MockCommunicator)
	state.Put("communicator", comm)
	state.Put("image_file_name", "foo")

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
	if !strings.Contains(comm.StartCmd.Command, "gsutil cp") {
		t.Fatalf("bad command: %#v", comm.StartCmd.Command)
	}
}

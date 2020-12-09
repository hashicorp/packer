package common

import (
	"context"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

func TestStepUploadParallelsTools_impl(t *testing.T) {
	var _ multistep.Step = new(StepUploadParallelsTools)
}

func TestStepUploadParallelsTools(t *testing.T) {
	state := testState(t)
	state.Put("parallels_tools_path", "./step_upload_parallels_tools_test.go")
	step := new(StepUploadParallelsTools)
	step.ParallelsToolsMode = "upload"
	step.ParallelsToolsGuestPath = "/tmp/prl-lin.iso"
	step.ParallelsToolsFlavor = "lin"

	comm := new(packersdk.MockCommunicator)
	state.Put("communicator", comm)

	// Test the run
	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	// Verify
	if comm.UploadPath != "/tmp/prl-lin.iso" {
		t.Fatalf("bad: %#v", comm.UploadPath)
	}
}

func TestStepUploadParallelsTools_interpolate(t *testing.T) {
	state := testState(t)
	state.Put("parallels_tools_path", "./step_upload_parallels_tools_test.go")
	step := new(StepUploadParallelsTools)
	step.ParallelsToolsMode = "upload"
	step.ParallelsToolsGuestPath = "/tmp/prl-{{ .Flavor }}.iso"
	step.ParallelsToolsFlavor = "win"

	comm := new(packersdk.MockCommunicator)
	state.Put("communicator", comm)

	// Test the run
	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	// Verify
	if comm.UploadPath != "/tmp/prl-win.iso" {
		t.Fatalf("bad: %#v", comm.UploadPath)
	}
}

func TestStepUploadParallelsTools_attach(t *testing.T) {
	state := testState(t)
	state.Put("parallels_tools_path", "./step_upload_parallels_tools_test.go")
	step := new(StepUploadParallelsTools)
	step.ParallelsToolsMode = "attach"
	step.ParallelsToolsGuestPath = "/tmp/prl-lin.iso"
	step.ParallelsToolsFlavor = "lin"

	comm := new(packersdk.MockCommunicator)
	state.Put("communicator", comm)

	// Test the run
	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	// Verify
	if comm.UploadCalled {
		t.Fatal("bad")
	}
}

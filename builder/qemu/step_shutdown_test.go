package qemu

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

func Test_Shutdown_Null_success(t *testing.T) {
	state := new(multistep.BasicStateBag)
	state.Put("ui", packersdk.TestUi(t))
	driverMock := new(DriverMock)
	driverMock.WaitForShutdownState = true
	state.Put("driver", driverMock)

	step := &stepShutdown{
		ShutdownCommand: "",
		ShutdownTimeout: 5 * time.Minute,
		Comm: &communicator.Config{
			Type: "none",
		},
	}
	action := step.Run(context.TODO(), state)
	if action != multistep.ActionContinue {
		t.Fatalf("Should have successfully shut down.")
	}
	err := state.Get("error")
	if err != nil {
		err = err.(error)
		t.Fatalf("Shutdown shouldn't have errored; err: %v", err)
	}
}

func Test_Shutdown_Null_failure(t *testing.T) {
	state := new(multistep.BasicStateBag)
	state.Put("ui", packersdk.TestUi(t))
	driverMock := new(DriverMock)
	driverMock.WaitForShutdownState = false
	state.Put("driver", driverMock)

	step := &stepShutdown{
		ShutdownCommand: "",
		ShutdownTimeout: 5 * time.Minute,
		Comm: &communicator.Config{
			Type: "none",
		},
	}
	action := step.Run(context.TODO(), state)
	if action != multistep.ActionHalt {
		t.Fatalf("Shouldn't have successfully shut down.")
	}
	err := state.Get("error")
	if err == nil {
		t.Fatalf("Shutdown should have errored")
	}
}

func Test_Shutdown_NoShutdownCommand(t *testing.T) {
	state := new(multistep.BasicStateBag)
	state.Put("ui", packersdk.TestUi(t))
	driverMock := new(DriverMock)
	state.Put("driver", driverMock)

	step := &stepShutdown{
		ShutdownCommand: "",
		ShutdownTimeout: 5 * time.Minute,
		Comm: &communicator.Config{
			Type: "ssh",
		},
	}
	action := step.Run(context.TODO(), state)
	if action != multistep.ActionContinue {
		t.Fatalf("Should have successfully shut down.")
	}

	if !driverMock.StopCalled {
		t.Fatalf("should have called Stop through the driver.")
	}
	err := state.Get("error")
	if err != nil {
		err = err.(error)
		t.Fatalf("Shutdown shouldn't have errored; err: %v", err)
	}
}

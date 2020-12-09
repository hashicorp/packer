package googlecompute

import (
	"context"
	"errors"
	"io/ioutil"
	"os"

	"github.com/hashicorp/packer-plugin-sdk/multistep"

	"testing"
)

func TestStepCreateOrResetWindowsPassword(t *testing.T) {
	state := testState(t)

	// Step is run after the instance is created so we will have an instance name set
	state.Put("instance_name", "mock_instance")
	state.Put("create_windows_password", true)

	step := new(StepCreateWindowsPassword)
	defer step.Cleanup(state)

	// run the step
	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}

	if password, ok := state.GetOk("winrm_password"); !ok || password.(string) != "MOCK_PASSWORD" {
		t.Fatal("should have a password", password, ok)
	}
}

func TestStepCreateOrResetWindowsPassword_passwordSet(t *testing.T) {
	state := testState(t)

	// Step is run after the instance is created so we will have an instance name set
	state.Put("instance_name", "mock_instance")

	c := state.Get("config").(*Config)

	c.Comm.WinRMPassword = "password"

	step := new(StepCreateWindowsPassword)
	defer step.Cleanup(state)

	// run the step
	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}

	if password, ok := state.GetOk("winrm_password"); !ok || password.(string) != "password" {
		t.Fatal("should have used existing password", password, ok)
	}
}

func TestStepCreateOrResetWindowsPassword_dontNeedPassword(t *testing.T) {
	state := testState(t)

	// Step is run after the instance is created so we will have an instance name set
	state.Put("instance_name", "mock_instance")

	step := new(StepCreateWindowsPassword)
	defer step.Cleanup(state)

	// run the step
	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}

}

func TestStepCreateOrResetWindowsPassword_debug(t *testing.T) {
	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(tf.Name())
	tf.Close()

	state := testState(t)
	// Step is run after the instance is created so we will have an instance name set
	state.Put("instance_name", "mock_instance")
	state.Put("create_windows_password", true)

	step := new(StepCreateWindowsPassword)

	step.Debug = true
	step.DebugKeyPath = tf.Name()

	defer step.Cleanup(state)

	// run the step
	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}

	if password, ok := state.GetOk("winrm_password"); !ok || password.(string) != "MOCK_PASSWORD" {
		t.Fatal("should have a password", password, ok)
	}

	if _, err := os.Stat(tf.Name()); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestStepCreateOrResetWindowsPassword_error(t *testing.T) {
	state := testState(t)

	// Step is run after the instance is created so we will have an instance name set
	state.Put("instance_name", "mock_instance")
	state.Put("create_windows_password", true)

	step := new(StepCreateWindowsPassword)
	defer step.Cleanup(state)

	driver := state.Get("driver").(*DriverMock)
	driver.CreateOrResetWindowsPasswordErr = errors.New("error")

	// run the step
	if action := step.Run(context.Background(), state); action != multistep.ActionHalt {
		t.Fatalf("bad action: %#v", action)
	}

	// Verify state
	if _, ok := state.GetOk("error"); !ok {
		t.Fatal("should have error")
	}

	if _, ok := state.GetOk("winrm_password"); ok {
		t.Fatal("should NOT have instance name")
	}
}

func TestStepCreateOrResetWindowsPassword_errorOnChannel(t *testing.T) {
	state := testState(t)

	// Step is run after the instance is created so we will have an instance name set
	state.Put("instance_name", "mock_instance")
	state.Put("create_windows_password", true)

	step := new(StepCreateWindowsPassword)
	defer step.Cleanup(state)

	driver := state.Get("driver").(*DriverMock)

	errCh := make(chan error, 1)
	errCh <- errors.New("error")

	driver.CreateOrResetWindowsPasswordErrCh = errCh

	// run the step
	if action := step.Run(context.Background(), state); action != multistep.ActionHalt {
		t.Fatalf("bad action: %#v", action)
	}

	// Verify state
	if _, ok := state.GetOk("error"); !ok {
		t.Fatal("should have error")
	}
	if _, ok := state.GetOk("winrm_password"); ok {
		t.Fatal("should NOT have instance name")
	}
}

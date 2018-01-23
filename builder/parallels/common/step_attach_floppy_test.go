package common

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/hashicorp/packer/helper/multistep"
)

func TestStepAttachFloppy_impl(t *testing.T) {
	var _ multistep.Step = new(StepAttachFloppy)
}

func TestStepAttachFloppy(t *testing.T) {
	state := testState(t)
	step := new(StepAttachFloppy)

	// Create a temporary file for our floppy file
	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	tf.Close()
	defer os.Remove(tf.Name())

	state.Put("floppy_path", tf.Name())
	state.Put("vmName", "foo")

	driver := state.Get("driver").(*DriverMock)

	// Test the run
	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	if len(driver.PrlctlCalls) != 2 {
		t.Fatal("not enough calls to prlctl")
	}

	if driver.PrlctlCalls[0][0] != "set" {
		t.Fatal("bad call")
	}
	if driver.PrlctlCalls[0][2] != "--device-del" {
		t.Fatal("bad call")
	}
	if driver.PrlctlCalls[0][3] != "fdd0" {
		t.Fatal("bad call")
	}

	if driver.PrlctlCalls[1][0] != "set" {
		t.Fatal("bad call")
	}
	if driver.PrlctlCalls[1][2] != "--device-add" {
		t.Fatal("bad call")
	}
	if driver.PrlctlCalls[1][3] != "fdd" {
		t.Fatal("bad call")
	}
	if driver.PrlctlCalls[1][6] != "--connect" {
		t.Fatal("bad call")
	}
}

func TestStepAttachFloppy_noFloppy(t *testing.T) {
	state := testState(t)
	step := new(StepAttachFloppy)

	driver := state.Get("driver").(*DriverMock)

	// Test the run
	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	if len(driver.PrlctlCalls) > 0 {
		t.Fatal("should not call prlctl")
	}
}

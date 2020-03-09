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

	if driver.RemoveFloppyControllersVM == "" {
		t.Fatal("RemoveFloppyControllers was not called")
	}

	if len(driver.VBoxManageCalls) != 2 {
		t.Fatal("not enough calls to VBoxManage")
	}
	if driver.VBoxManageCalls[0][0] != "storagectl" {
		t.Fatal("bad call")
	}
	if driver.VBoxManageCalls[1][0] != "storageattach" {
		t.Fatal("bad call")
	}

	// Test the cleanup
	step.Cleanup(state)
	if driver.VBoxManageCalls[2][0] != "storageattach" {
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

	if len(driver.VBoxManageCalls) > 0 {
		t.Fatal("should not call vboxmanage")
	}
}

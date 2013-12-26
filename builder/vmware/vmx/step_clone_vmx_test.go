package vmx

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/mitchellh/multistep"
)

func TestStepCloneVMX_impl(t *testing.T) {
	var _ multistep.Step = new(StepCloneVMX)
}

func TestStepCloneVMX(t *testing.T) {
	// Setup some state
	td, err := ioutil.TempDir("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.RemoveAll(td)

	// Create the source
	sourcePath := filepath.Join(td, "source.vmx")
	if err := ioutil.WriteFile(sourcePath, []byte("foo"), 0644); err != nil {
		t.Fatalf("err: %s", err)
	}

	state := testState(t)
	step := new(StepCloneVMX)
	step.OutputDir = td
	step.Path = sourcePath
	step.VMName = "foo"

	// Test the run
	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	// Test we have our VMX
	if _, err := os.Stat(filepath.Join(td, "foo.vmx")); err != nil {
		t.Fatalf("err: %s", err)
	}

	data, err := ioutil.ReadFile(filepath.Join(td, "foo.vmx"))
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if string(data) != "foo" {
		t.Fatalf("bad: %#v", string(data))
	}
}

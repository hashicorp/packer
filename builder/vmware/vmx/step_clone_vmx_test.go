package vmx

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/mitchellh/multistep"
	vmwcommon "github.com/mitchellh/packer/builder/vmware/common"
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
	if err := ioutil.WriteFile(sourcePath, []byte(testCloneVMX), 0644); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Create the dest because the mock driver won't
	destPath := filepath.Join(td, "foo.vmx")
	if err := ioutil.WriteFile(destPath, []byte(testCloneVMX), 0644); err != nil {
		t.Fatalf("err: %s", err)
	}

	state := testState(t)
	step := new(StepCloneVMX)
	step.OutputDir = td
	step.Path = sourcePath
	step.VMName = "foo"

	driver := state.Get("driver").(*vmwcommon.DriverMock)

	// Test the run
	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	// Test we cloned
	if !driver.CloneCalled {
		t.Fatal("should call clone")
	}

	// Test that we have our paths
	if vmxPath, ok := state.GetOk("vmx_path"); !ok {
		t.Fatal("should set vmx_path")
	} else if vmxPath != destPath {
		t.Fatalf("bad: %#v", vmxPath)
	}

	if diskPath, ok := state.GetOk("full_disk_path"); !ok {
		t.Fatal("should set full_disk_path")
	} else if diskPath != filepath.Join(td, "foo") {
		t.Fatalf("bad: %#v", diskPath)
	}
}

const testCloneVMX = `
scsi0:0.fileName = "foo"
`

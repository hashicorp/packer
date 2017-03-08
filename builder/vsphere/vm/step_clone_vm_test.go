package vm

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/mitchellh/multistep"
	vspcommon "github.com/mitchellh/packer/builder/vsphere/common"
)

func TestStepCloneVM_impl(t *testing.T) {
	var _ multistep.Step = new(stepCloneVM)
}

func TestStepCloneVM(t *testing.T) {
	// Setup some state
	td, err := ioutil.TempDir("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.RemoveAll(td)

	state := testState(t)
	var config Config
	state.Put("config", &config)

	step := new(stepCloneVM)
	step.VMName = "foo"
	step.SrcVMName = "src-foo"
	step.Folder = "fold-foo"
	step.Datastore = "fold-data"
	step.Cpu = 3
	step.MemSize = 256
	step.DiskSize = 50000
	step.DiskThick = true
	step.NetworkName = "net-foo"
	step.NetworkAdapter = "vmxnet3"
	step.Annotation = "foobar"

	driver := state.Get("driver").(*vspcommon.DriverMock)

	// Test the run
	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	// Test we cloned
	if !driver.CloneVirtualMachineCalled {
		t.Fatal("should call clone")
	}
}

func TestStepCloneVM_AdditionnalDisk(t *testing.T) {
	// Setup some state
	td, err := ioutil.TempDir("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.RemoveAll(td)

	state := testState(t)
	var config Config
	config.AdditionalDiskSize = []uint{2, 3, 1}
	state.Put("config", &config)

	step := new(stepCloneVM)
	step.VMName = "foo"
	step.SrcVMName = "src-foo"
	step.Folder = "fold-foo"
	step.Datastore = "fold-data"
	step.Cpu = 3
	step.MemSize = 256
	step.DiskSize = 50000
	step.DiskThick = true
	step.NetworkName = "net-foo"
	step.NetworkAdapter = "vmxnet3"
	step.Annotation = "foobar"

	driver := state.Get("driver").(*vspcommon.DriverMock)

	// Test the run
	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	// Test we cloned
	if !driver.CloneVirtualMachineCalled {
		t.Fatal("should call clone")
	}

	if !driver.CreateDiskCalled {
		t.Fatal("should call clone")
	}

	created := driver.CreateDiskOutput
	if created[0] != 2 || created[1] != 3 || created[2] != 1 || len(created) != 3 {
		t.Fatalf("Additional disk size incorrectly processed: %#v", driver.CreateDiskOutput)
	}
}

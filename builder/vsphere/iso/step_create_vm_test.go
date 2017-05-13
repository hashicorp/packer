package iso

import (
	"io/ioutil"
	"os"
	"testing"

	vspcommon "github.com/hashicorp/packer/builder/vsphere/common"
	"github.com/mitchellh/multistep"
)

func TestStepCreateVM_impl(t *testing.T) {
	var _ multistep.Step = new(stepCreateVM)
}

func TestStepCreateVM(t *testing.T) {
	// Setup some state
	td, err := ioutil.TempDir("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.RemoveAll(td)

	state := testState(t)
	var config Config
	state.Put("config", &config)
	state.Put("iso_path", "foocdrom")

	step := new(stepCreateVM)
	step.VMName = "foo"
	step.Folder = "fold-foo"
	step.Datastore = "fold-data"
	step.GuestType = "other16"
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

	if !driver.CreateVirtualMachineCalled {
		t.Fatal("should call create")
	}

	if driver.AddFloppyCalled {
		t.Fatal("should not call add floppy")
	}

	if !driver.MountISOCalled {
		t.Fatal("should call mount iso")
	}

	cdrom_device := state.Get("cdrom_device")
	if cdrom_device != "cdrom1" {
		t.Fatalf("wrong cdrom device: %s", cdrom_device)
	}
}

func TestStepCreateVMFloppy(t *testing.T) {
	// Setup some state
	td, err := ioutil.TempDir("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.RemoveAll(td)

	state := testState(t)
	var config Config
	state.Put("config", &config)
	state.Put("iso_path", "foocdrom")
	state.Put("floppy_path", "foofloppy")

	step := new(stepCreateVM)
	step.VMName = "foo"
	step.Folder = "fold-foo"
	step.Datastore = "fold-data"
	step.GuestType = "other16"
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

	if !driver.CreateVirtualMachineCalled {
		t.Fatal("should call create")
	}

	if !driver.AddFloppyCalled {
		t.Fatal("should call add floppy")
	}

	floppy_device := state.Get("floppy_device")
	if floppy_device != "floppy1" {
		t.Fatalf("wrong floppy device: %s", floppy_device)
	}

	if !driver.MountISOCalled {
		t.Fatal("should call mount iso")
	}

	cdrom_device := state.Get("cdrom_device")
	if cdrom_device != "cdrom1" {
		t.Fatalf("wrong cdrom device: %s", cdrom_device)
	}
}

func TestStepCreateVM_AdditionnalDisk(t *testing.T) {
	// Setup some state
	td, err := ioutil.TempDir("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.RemoveAll(td)

	state := testState(t)
	var config Config
	state.Put("config", &config)
	state.Put("iso_path", "foocdrom")

	step := new(stepCreateVM)
	step.VMName = "foo"
	step.GuestType = "other16"
	step.Folder = "fold-foo"
	step.Datastore = "fold-data"
	step.Cpu = 3
	step.MemSize = 256
	step.DiskSize = 50000
	step.AdditionalDiskSize = []uint{2, 3, 1}
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

	if !driver.CreateVirtualMachineCalled {
		t.Fatal("should call create")
	}

	if !driver.CreateDiskCalled {
		t.Fatal("should call create")
	}

	created := driver.CreateDiskOutput
	if created[0] != 2 || created[1] != 3 || created[2] != 1 || len(created) != 3 {
		t.Fatalf("Additional disk size incorrectly processed: %#v", driver.CreateDiskOutput)
	}
}

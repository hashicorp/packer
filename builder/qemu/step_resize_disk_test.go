package qemu

import (
	"context"
	"testing"

	"github.com/hashicorp/packer/helper/multistep"
)

func TestStepResizeDisk_Run(t *testing.T) {
	state := testState(t)
	driver := state.Get("driver").(*DriverMock)

	config := &Config{
		DiskImage:      true,
		SkipResizeDisk: false,
		DiskSize:       "4096M",
		Format:         "qcow2",
		OutputDir:      "/test/",
		VMName:         "test",
	}
	state.Put("config", config)
	step := new(stepResizeDisk)

	// Test the run
	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}
	if len(driver.QemuImgCalls) == 0 {
		t.Fatal("should qemu-img called")
	}
	if len(driver.QemuImgCalls[0]) != 5 {
		t.Fatal("should 5 qemu-img parameters")
	}
}

func TestStepResizeDisk_SkipIso(t *testing.T) {
	state := testState(t)
	driver := state.Get("driver").(*DriverMock)
	config := &Config{
		DiskImage:      false,
		SkipResizeDisk: false,
	}
	state.Put("config", config)
	step := new(stepResizeDisk)

	// Test the run
	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}
	if len(driver.QemuImgCalls) > 0 {
		t.Fatal("should NOT qemu-img called")
	}
}

func TestStepResizeDisk_SkipOption(t *testing.T) {
	state := testState(t)
	driver := state.Get("driver").(*DriverMock)
	config := &Config{
		DiskImage:      false,
		SkipResizeDisk: true,
	}
	state.Put("config", config)
	step := new(stepResizeDisk)

	// Test the run
	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}
	if len(driver.QemuImgCalls) > 0 {
		t.Fatal("should NOT qemu-img called")
	}
}

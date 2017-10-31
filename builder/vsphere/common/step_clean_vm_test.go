package common

import (
	"sort"
	"testing"

	"github.com/mitchellh/multistep"
)

func TestStepCleanVM_impl(t *testing.T) {
	var _ multistep.Step = new(StepCleanVM)
}

func TestStepCleanVM_ChangeFloppyCdrom(t *testing.T) {
	state := testState(t)
	driver := new(DriverMock)
	state.Put("driver", driver)
	state.Put("floppy_device", "foofloppy")
	state.Put("cdrom_device", "foocdrom")

	step := new(StepCleanVM)
	step.CustomData = map[string]string{"foo": "bar", "foo1": "bar1"}

	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	if !driver.RemoveFloppyCalled {
		t.Fatal("remove floppy should be called")
	}

	if !driver.VMChangeCalled {
		t.Fatal("VM change should be called")
	}

	if !driver.UnmountISOCalled {
		t.Fatal("unmount ISO should be called")
	}

	if !driver.VNCDisableCalled {
		t.Fatal("VNC disable should be called")
	}

	vco := driver.VMChangeOption
	sort.Strings(vco)
	if vco[0] != "foo1=bar1" || vco[1] != "foo=bar" || len(vco) != 2 {
		t.Fatalf("wrong return from VM change: %#v", vco)
	}

	// Cleanup
	step.Cleanup(state)
}

func TestStepCleanVM_FloppyCdrom(t *testing.T) {
	state := testState(t)
	driver := new(DriverMock)
	state.Put("driver", driver)
	state.Put("floppy_device", "foofloppy")
	state.Put("cdrom_device", "foocdrom")

	step := new(StepCleanVM)

	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	if !driver.RemoveFloppyCalled {
		t.Fatal("remove floppy should be called")
	}

	if driver.VMChangeCalled {
		t.Fatal("VM change should not be called")
	}

	if !driver.UnmountISOCalled {
		t.Fatal("unmount ISO should be called")
	}

	if !driver.VNCDisableCalled {
		t.Fatal("VNC disable should be called")
	}

	// Cleanup
	step.Cleanup(state)
}

func TestStepCleanVM_ChangeCdrom(t *testing.T) {
	state := testState(t)
	driver := new(DriverMock)
	state.Put("driver", driver)
	state.Put("cdrom_device", "foocdrom")

	step := new(StepCleanVM)
	step.CustomData = map[string]string{"foo": "bar", "foo1": "bar1"}

	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	if driver.RemoveFloppyCalled {
		t.Fatal("remove floppy should not be called")
	}

	if !driver.VMChangeCalled {
		t.Fatal("VM change should be called")
	}

	if !driver.UnmountISOCalled {
		t.Fatal("unmount ISO should be called")
	}

	if !driver.VNCDisableCalled {
		t.Fatal("VNC disable should be called")
	}

	// Cleanup
	step.Cleanup(state)
}

func TestStepCleanVM_ChangeFloppy(t *testing.T) {
	state := testState(t)
	driver := new(DriverMock)
	state.Put("driver", driver)
	state.Put("floppy_device", "foofloppy")

	step := new(StepCleanVM)
	step.CustomData = map[string]string{"foo": "bar", "foo1": "bar1"}

	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	if !driver.RemoveFloppyCalled {
		t.Fatal("remove floppy should be called")
	}

	if !driver.VMChangeCalled {
		t.Fatal("VM change should be called")
	}

	if driver.UnmountISOCalled {
		t.Fatal("unmount ISO should not be called")
	}

	if !driver.VNCDisableCalled {
		t.Fatal("VNC disable should be called")
	}

	// Cleanup
	step.Cleanup(state)
}

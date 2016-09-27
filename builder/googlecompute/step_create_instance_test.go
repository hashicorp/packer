package googlecompute

import (
	"errors"
	"testing"
	"time"

	"github.com/mitchellh/multistep"
	"github.com/stretchr/testify/assert"
)

func TestStepCreateInstance_impl(t *testing.T) {
	var _ multistep.Step = new(StepCreateInstance)
}

func TestStepCreateInstance(t *testing.T) {
	state := testState(t)
	step := new(StepCreateInstance)
	defer step.Cleanup(state)

	state.Put("ssh_public_key", "key")

	c := state.Get("config").(*Config)
	d := state.Get("driver").(*DriverMock)
	d.GetImageResult = StubImage("test-image", "test-project", []string{}, 100)

	// run the step
	assert.Equal(t, step.Run(state), multistep.ActionContinue, "Step should have passed and continued.")

	// Verify state
	nameRaw, ok := state.GetOk("instance_name")
	assert.True(t, ok, "State should have an instance name.")

	// cleanup
	step.Cleanup(state)

	// Check args passed to the driver.
	assert.Equal(t, d.DeleteInstanceName, nameRaw.(string), "Incorrect instance name passed to driver.")
	assert.Equal(t, d.DeleteInstanceZone, c.Zone, "Incorrect instance zone passed to driver.")
	assert.Equal(t, d.DeleteDiskName, c.InstanceName, "Incorrect disk name passed to driver.")
	assert.Equal(t, d.DeleteDiskZone, c.Zone, "Incorrect disk zone passed to driver.")
}

func TestStepCreateInstance_windowsNeedsPassword(t *testing.T) {

	state := testState(t)
	step := new(StepCreateInstance)
	defer step.Cleanup(state)

	state.Put("ssh_public_key", "key")
	c := state.Get("config").(*Config)
	d := state.Get("driver").(*DriverMock)
	d.GetImageResult = StubImage("test-image", "test-project", []string{"windows"}, 100)
	c.Comm.Type = "winrm"
	// run the step
	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}

	// Verify state
	nameRaw, ok := state.GetOk("instance_name")
	if !ok {
		t.Fatal("should have instance name")
	}

	createPassword, ok := state.GetOk("create_windows_password")

	if !ok || !createPassword.(bool) {
		t.Fatal("should need to create a windows password")
	}

	// cleanup
	step.Cleanup(state)

	if d.DeleteInstanceName != nameRaw.(string) {
		t.Fatal("should've deleted instance")
	}
	if d.DeleteInstanceZone != c.Zone {
		t.Fatalf("bad instance zone: %#v", d.DeleteInstanceZone)
	}

	if d.DeleteDiskName != c.InstanceName {
		t.Fatal("should've deleted disk")
	}
	if d.DeleteDiskZone != c.Zone {
		t.Fatalf("bad disk zone: %#v", d.DeleteDiskZone)
	}
}

func TestStepCreateInstance_windowsPasswordSet(t *testing.T) {

	state := testState(t)
	step := new(StepCreateInstance)
	defer step.Cleanup(state)

	state.Put("ssh_public_key", "key")

	config := state.Get("config").(*Config)
	driver := state.Get("driver").(*DriverMock)
	driver.GetImageResult = StubImage("test-image", "test-project", []string{"windows"}, 100)
	config.Comm.Type = "winrm"
	config.Comm.WinRMPassword = "password"

	// run the step
	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}

	// Verify state
	nameRaw, ok := state.GetOk("instance_name")
	if !ok {
		t.Fatal("should have instance name")
	}

	_, ok = state.GetOk("create_windows_password")

	if ok {
		t.Fatal("should not need to create windows password")
	}

	// cleanup
	step.Cleanup(state)

	if driver.DeleteInstanceName != nameRaw.(string) {
		t.Fatal("should've deleted instance")
	}
	if driver.DeleteInstanceZone != config.Zone {
		t.Fatalf("bad instance zone: %#v", driver.DeleteInstanceZone)
	}

	if driver.DeleteDiskName != config.InstanceName {
		t.Fatal("should've deleted disk")
	}
	if driver.DeleteDiskZone != config.Zone {
		t.Fatalf("bad disk zone: %#v", driver.DeleteDiskZone)
	}
}

func TestStepCreateInstance_error(t *testing.T) {
	state := testState(t)
	step := new(StepCreateInstance)
	defer step.Cleanup(state)

	state.Put("ssh_public_key", "key")

	d := state.Get("driver").(*DriverMock)
	d.RunInstanceErr = errors.New("error")
	d.GetImageResult = StubImage("test-image", "test-project", []string{}, 100)

	// run the step
	assert.Equal(t, step.Run(state), multistep.ActionHalt, "Step should have failed and halted.")

	// Verify state
	_, ok := state.GetOk("error")
	assert.True(t, ok, "State should have an error.")
	_, ok = state.GetOk("instance_name")
	assert.False(t, ok, "State should not have an instance name.")
}

func TestStepCreateInstance_errorOnChannel(t *testing.T) {
	state := testState(t)
	step := new(StepCreateInstance)
	defer step.Cleanup(state)

	state.Put("ssh_public_key", "key")

	errCh := make(chan error, 1)
	errCh <- errors.New("error")

	d := state.Get("driver").(*DriverMock)
	d.RunInstanceErrCh = errCh
	d.GetImageResult = StubImage("test-image", "test-project", []string{}, 100)

	// run the step
	assert.Equal(t, step.Run(state), multistep.ActionHalt, "Step should have failed and halted.")

	// Verify state
	_, ok := state.GetOk("error")
	assert.True(t, ok, "State should have an error.")
	_, ok = state.GetOk("instance_name")
	assert.False(t, ok, "State should not have an instance name.")
}

func TestStepCreateInstance_errorTimeout(t *testing.T) {
	state := testState(t)
	step := new(StepCreateInstance)
	defer step.Cleanup(state)

	state.Put("ssh_public_key", "key")

	errCh := make(chan error, 1)
	go func() {
		<-time.After(10 * time.Millisecond)
		errCh <- nil
	}()

	config := state.Get("config").(*Config)
	config.stateTimeout = 1 * time.Microsecond

	d := state.Get("driver").(*DriverMock)
	d.RunInstanceErrCh = errCh
	d.GetImageResult = StubImage("test-image", "test-project", []string{}, 100)

	// run the step
	assert.Equal(t, step.Run(state), multistep.ActionHalt, "Step should have failed and halted.")

	// Verify state
	_, ok := state.GetOk("error")
	assert.True(t, ok, "State should have an error.")
	_, ok = state.GetOk("instance_name")
	assert.False(t, ok, "State should not have an instance name.")
}

package common

import (
	"testing"
	"time"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

func TestStepShutdown_impl(t *testing.T) {
	var _ multistep.Step = new(StepShutdown)
}

func TestStepShutdown_noShutdownCommand(t *testing.T) {
	state := testState(t)
	step := new(StepShutdown)

	comm := new(packer.MockCommunicator)
	state.Put("communicator", comm)
	state.Put("vmName", "foo")

	driver := state.Get("driver").(*DriverMock)

	// Test the run
	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	// Test that Stop was just called
	if driver.StopName != "foo" {
		t.Fatal("should call stop")
	}
	if comm.StartCalled {
		t.Fatal("comm start should not be called")
	}
}

func TestStepShutdown_shutdownCommand(t *testing.T) {
	state := testState(t)
	step := new(StepShutdown)
	step.Command = "poweroff"
	step.Timeout = 1 * time.Second

	comm := new(packer.MockCommunicator)
	state.Put("communicator", comm)
	state.Put("vmName", "foo")

	driver := state.Get("driver").(*DriverMock)
	driver.IsRunningReturn = true

	go func() {
		time.Sleep(10 * time.Millisecond)
		driver.Lock()
		defer driver.Unlock()
		driver.IsRunningReturn = false
	}()

	// Test the run
	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	// Test that Stop was just called
	if driver.StopName != "" {
		t.Fatal("should not call stop")
	}
	if comm.StartCmd.Command != step.Command {
		t.Fatal("comm start should be called")
	}
}

func TestStepShutdown_shutdownTimeout(t *testing.T) {
	state := testState(t)
	step := new(StepShutdown)
	step.Command = "poweroff"
	step.Timeout = 1 * time.Second

	comm := new(packer.MockCommunicator)
	state.Put("communicator", comm)
	state.Put("vmName", "foo")

	driver := state.Get("driver").(*DriverMock)
	driver.IsRunningReturn = true

	go func() {
		time.Sleep(2 * time.Second)
		driver.Lock()
		defer driver.Unlock()
		driver.IsRunningReturn = false
	}()

	// Test the run
	if action := step.Run(state); action != multistep.ActionHalt {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); !ok {
		t.Fatal("should have error")
	}
}

func TestStepShutdown_shutdownDelay(t *testing.T) {
	state := testState(t)
	step := new(StepShutdown)
	step.Command = "poweroff"
	step.Timeout = 5 * time.Second
	step.Delay = 2 * time.Second

	comm := new(packer.MockCommunicator)
	state.Put("communicator", comm)
	state.Put("vmName", "foo")

	driver := state.Get("driver").(*DriverMock)
	driver.IsRunningReturn = true
	start := time.Now()

	go func() {
		time.Sleep(10 * time.Millisecond)
		driver.Lock()
		defer driver.Unlock()
		driver.IsRunningReturn = false
	}()

	// Test the run

	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	testDuration := time.Since(start).Seconds()
	if testDuration < 2.5 || testDuration > 2.6 {
		t.Fatal("incorrect duration")
	}

	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	step.Delay = 0

	driver.IsRunningReturn = true
	start = time.Now()

	go func() {
		time.Sleep(10 * time.Millisecond)
		driver.Lock()
		defer driver.Unlock()
		driver.IsRunningReturn = false
	}()

	// Test the run
	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	testDuration = time.Since(start).Seconds()
	if testDuration > 0.6 {
		t.Fatal("incorrect duration")
	}

	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

}

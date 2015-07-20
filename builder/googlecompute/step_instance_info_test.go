package googlecompute

import (
	"errors"
	"github.com/mitchellh/multistep"
	"testing"
	"time"
)

func TestStepInstanceInfo_impl(t *testing.T) {
	var _ multistep.Step = new(StepInstanceInfo)
}

func TestStepInstanceInfo(t *testing.T) {
	state := testState(t)
	step := new(StepInstanceInfo)
	defer step.Cleanup(state)

	state.Put("instance_name", "foo")

	config := state.Get("config").(*Config)
	driver := state.Get("driver").(*DriverMock)
	driver.GetNatIPResult = "1.2.3.4"

	// run the step
	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}

	// Verify state
	if driver.WaitForInstanceState != "RUNNING" {
		t.Fatalf("bad: %#v", driver.WaitForInstanceState)
	}
	if driver.WaitForInstanceZone != config.Zone {
		t.Fatalf("bad: %#v", driver.WaitForInstanceZone)
	}
	if driver.WaitForInstanceName != "foo" {
		t.Fatalf("bad: %#v", driver.WaitForInstanceName)
	}

	ipRaw, ok := state.GetOk("instance_ip")
	if !ok {
		t.Fatal("should have ip")
	}
	if ip, ok := ipRaw.(string); !ok {
		t.Fatal("ip is not a string")
	} else if ip != "1.2.3.4" {
		t.Fatalf("bad ip: %s", ip)
	}
}

func TestStepInstanceInfo_InternalIP(t *testing.T) {
	state := testState(t)
	step := new(StepInstanceInfo)
	defer step.Cleanup(state)

	state.Put("instance_name", "foo")

	config := state.Get("config").(*Config)
	config.UseInternalIP = true
	driver := state.Get("driver").(*DriverMock)
	driver.GetNatIPResult = "1.2.3.4"
	driver.GetInternalIPResult = "5.6.7.8"

	// run the step
	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}

	// Verify state
	if driver.WaitForInstanceState != "RUNNING" {
		t.Fatalf("bad: %#v", driver.WaitForInstanceState)
	}
	if driver.WaitForInstanceZone != config.Zone {
		t.Fatalf("bad: %#v", driver.WaitForInstanceZone)
	}
	if driver.WaitForInstanceName != "foo" {
		t.Fatalf("bad: %#v", driver.WaitForInstanceName)
	}

	ipRaw, ok := state.GetOk("instance_ip")
	if !ok {
		t.Fatal("should have ip")
	}
	if ip, ok := ipRaw.(string); !ok {
		t.Fatal("ip is not a string")
	} else if ip != "5.6.7.8" {
		t.Fatalf("bad ip: %s", ip)
	}
}

func TestStepInstanceInfo_getNatIPError(t *testing.T) {
	state := testState(t)
	step := new(StepInstanceInfo)
	defer step.Cleanup(state)

	state.Put("instance_name", "foo")

	driver := state.Get("driver").(*DriverMock)
	driver.GetNatIPErr = errors.New("error")

	// run the step
	if action := step.Run(state); action != multistep.ActionHalt {
		t.Fatalf("bad action: %#v", action)
	}

	// Verify state
	if _, ok := state.GetOk("error"); !ok {
		t.Fatal("should have error")
	}
	if _, ok := state.GetOk("instance_ip"); ok {
		t.Fatal("should NOT have instance IP")
	}
}

func TestStepInstanceInfo_waitError(t *testing.T) {
	state := testState(t)
	step := new(StepInstanceInfo)
	defer step.Cleanup(state)

	state.Put("instance_name", "foo")

	errCh := make(chan error, 1)
	errCh <- errors.New("error")

	driver := state.Get("driver").(*DriverMock)
	driver.WaitForInstanceErrCh = errCh

	// run the step
	if action := step.Run(state); action != multistep.ActionHalt {
		t.Fatalf("bad action: %#v", action)
	}

	// Verify state
	if _, ok := state.GetOk("error"); !ok {
		t.Fatal("should have error")
	}
	if _, ok := state.GetOk("instance_ip"); ok {
		t.Fatal("should NOT have instance IP")
	}
}

func TestStepInstanceInfo_errorTimeout(t *testing.T) {
	state := testState(t)
	step := new(StepInstanceInfo)
	defer step.Cleanup(state)

	errCh := make(chan error, 1)
	go func() {
		<-time.After(10 * time.Millisecond)
		errCh <- nil
	}()

	state.Put("instance_name", "foo")

	config := state.Get("config").(*Config)
	config.stateTimeout = 1 * time.Microsecond

	driver := state.Get("driver").(*DriverMock)
	driver.WaitForInstanceErrCh = errCh

	// run the step
	if action := step.Run(state); action != multistep.ActionHalt {
		t.Fatalf("bad action: %#v", action)
	}

	// Verify state
	if _, ok := state.GetOk("error"); !ok {
		t.Fatal("should have error")
	}
	if _, ok := state.GetOk("instance_ip"); ok {
		t.Fatal("should NOT have instance IP")
	}
}

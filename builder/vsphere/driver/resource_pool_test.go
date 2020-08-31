package driver

import (
	"testing"

	"github.com/vmware/govmomi/simulator"
)

func TestVCenterDriver_FindResourcePool(t *testing.T) {
	model := simulator.VPX()
	defer model.Remove()

	s, err := NewSimulatorServer(model)
	if err != nil {
		t.Fatalf("should not fail: %s", err.Error())
	}
	defer s.Close()

	driverSim, err := NewSimulatorDriver(s)
	if err != nil {
		t.Fatalf("should not fail: %s", err.Error())
	}

	res, err := driverSim.FindResourcePool("", "DC0_H0", "")
	if err != nil {
		t.Fatalf("should not fail: %s", err.Error())
	}
	if res == nil {
		t.Fatalf("resource pool should not be nil")
	}
	expectedResourcePool := "Resources"
	if res.pool.Name() != expectedResourcePool {
		t.Fatalf("resource name expected %s but was %s", expectedResourcePool, res.pool.Name())
	}
}

func TestVCenterDriver_FindResourcePoolStandaloneESX(t *testing.T) {
	// standalone ESX host without any vCenter
	model := simulator.ESX()
	defer model.Remove()

	opts := simulator.VPX()
	model.Datastore = opts.Datastore
	model.Machine = opts.Machine
	model.Autostart = opts.Autostart
	model.DelayConfig.Delay = opts.DelayConfig.Delay
	model.DelayConfig.MethodDelay = opts.DelayConfig.MethodDelay
	model.DelayConfig.DelayJitter = opts.DelayConfig.DelayJitter

	s, err := NewSimulatorServer(model)
	if err != nil {
		t.Fatalf("should not fail: %s", err.Error())
	}
	defer s.Close()

	driverSim, err := NewSimulatorDriver(s)
	if err != nil {
		t.Fatalf("should not fail: %s", err.Error())
	}

	//
	res, err := driverSim.FindResourcePool("", "localhost.localdomain", "")
	if err != nil {
		t.Fatalf("should not fail: %s", err.Error())
	}
	if res == nil {
		t.Fatalf("resource pool should not be nil")
	}
	expectedResourcePool := "Resources"
	if res.pool.Name() != expectedResourcePool {
		t.Fatalf("resource name expected %s but was %s", expectedResourcePool, res.pool.Name())
	}

	// Invalid resource name should look for default resource pool
	res, err = driverSim.FindResourcePool("", "localhost.localdomain", "invalid")
	if err != nil {
		t.Fatalf("should not fail: %s", err.Error())
	}
	if res == nil {
		t.Fatalf("resource pool should not be nil")
	}
	if res.pool.Name() != expectedResourcePool {
		t.Fatalf("resource name expected %s but was %s", expectedResourcePool, res.pool.Name())
	}
}

package driver

import (
	"testing"

	"github.com/vmware/govmomi/simulator"
)

func TestVCenterDriver_FindResourcePool(t *testing.T) {
	sim, err := NewVCenterSimulator()
	if err != nil {
		t.Fatalf("should not fail: %s", err.Error())
	}
	defer sim.Close()

	res, err := sim.driver.FindResourcePool("", "DC0_H0", "")
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

	sim, err := NewCustomVCenterSimulator(model)
	if err != nil {
		t.Fatalf("should not fail: %s", err.Error())
	}
	defer sim.Close()

	res, err := sim.driver.FindResourcePool("", "localhost.localdomain", "")
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
	res, err = sim.driver.FindResourcePool("", "localhost.localdomain", "invalid")
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

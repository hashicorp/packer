package testing

import (
	"github.com/jetbrains-infra/packer-builder-vsphere/driver"
	"log"
	"testing"
	"net"
	"time"
)

func initVMAccTest(t *testing.T) (d *driver.Driver, vm *driver.VirtualMachine, vmName string, vmDestructor func()) {
	initDriverAcceptanceTest(t)

	templateName := "alpine"
	d = NewTestDriver(t)

	template, err := d.FindVM(templateName) // Don't destroy this VM!
	if err != nil {
		t.Fatalf("Cannot find template vm '%v': %v", templateName, err)
	}

	log.Printf("[DEBUG] Clonning VM")
	vmName = NewVMName()
	vm, err = template.Clone(&driver.CloneConfig{
		Name: vmName,
		Host: "esxi-1.vsphere55.test",
	})
	if err != nil {
		t.Fatalf("Cannot clone vm '%v': %v", templateName, err)
	}

	vmDestructor = func() {
		log.Printf("[DEBUG] Removing the clone")
		if err := vm.Destroy(); err != nil {
			t.Errorf("!!! ERROR REMOVING VM '%v': %v!!!", vmName, err)
		}

		// Check that the clone is no longer exists
		if _, err := d.FindVM(vmName); err == nil {
			t.Errorf("!!! STILL CAN FIND VM '%v'. IT MIGHT NOT HAVE BEEN DELETED !!!", vmName)
		}
	}

	return
}

func TestVMAcc_default(t *testing.T) {
	d, vm, vmName, vmDestructor := initVMAccTest(t)
	defer vmDestructor()

	// Check that the clone can be found by its name
	if _, err := d.FindVM(vmName); err != nil {
		t.Errorf("Cannot find created vm '%v': %v", vmName, err)
	}

	// Run checks
	log.Printf("[DEBUG] Running check function")
	VMCheckDefault(t, d, vm, vmName, "esxi-1.vsphere55.test", "datastore1")
}

func TestVMAcc_hardware(t *testing.T) {
	d, vm, _ /*vmName*/, vmDestructor := initVMAccTest(t)
	defer vmDestructor()

	log.Printf("[DEBUG] Configuring the vm")
	vm.Configure(&driver.HardwareConfig{
		CPUs:           2,
		CPUReservation: 1000,
		CPULimit:       1500,
		RAM:            2048,
		RAMReservation: 1024,
	})
	log.Printf("[DEBUG] Running check function")
	VMCheckHardware(t, d, vm)
}

func startVM(t *testing.T, vm *driver.VirtualMachine, vmName string) (stopper func()) {
	log.Printf("[DEBUG] Starting the vm")
	if err := vm.PowerOn(); err != nil {
		t.Fatalf("Cannot start created vm '%v': %v", vmName, err)
	}
	return func() {
		log.Printf("[DEBUG] Powering off the vm")
		if err := vm.PowerOff(); err != nil {
			t.Errorf("Cannot power off started vm '%v': %v", vmName, err)
		}
	}
}

func TestVMAcc_running(t *testing.T) {
	_ /*d*/, vm, vmName, vmDestructor := initVMAccTest(t)
	defer vmDestructor()

	stopper := startVM(t, vm, vmName)
	defer stopper()

	switch ip, err := vm.WaitForIP(); {
	case err != nil:
		t.Errorf("Cannot obtain IP address from created vm '%v': %v", vmName, err)
	case net.ParseIP(ip) == nil:
		t.Errorf("'%v' is not a valid ip address", ip)
	}

	vm.StartShutdown()
	log.Printf("[DEBUG] Waiting max 1m0s for shutdown to complete")
	// TODO: there is complex logic in WaitForShutdown. It's better to test it well. It might be reasonable to create
	// unit tests for it.
	vm.WaitForShutdown(1 * time.Minute)
}

func TestVMAcc_running_snapshot(t *testing.T) {
	d, vm, vmName, vmDestructor := initVMAccTest(t)
	defer vmDestructor()

	stopper := startVM(t, vm, vmName)
	defer stopper()

	vm.CreateSnapshot("test-snapshot")
	VMCheckSnapshor(t, d, vm)
}

func TestVMAcc_template(t *testing.T) {
	d, vm, _ /*vmName*/, vmDestructor := initVMAccTest(t)
	defer vmDestructor()

	vm.ConvertToTemplate()
	VMCheckTemplate(t, d, vm)
}

package driver

import (
	"log"
	"testing"
	"net"
	"time"
)

func initVMAccTest(t *testing.T) (d *Driver, vm *VirtualMachine, vmName string, vmDestructor func()) {
	initDriverAcceptanceTest(t)

	templateName := "alpine"
	d = newTestDriver(t)

	template, err := d.FindVM(templateName) // Don't destroy this VM!
	if err != nil {
		t.Fatalf("Cannot find template vm '%v': %v", templateName, err)
	}

	log.Printf("[DEBUG] Clonning VM")
	vmName = newVMName()
	vm, err = template.Clone(&CloneConfig{
		Name: vmName,
		Host: hostName,
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
	log.Printf("[DEBUG] Running checks")
	vmInfo, err := vm.Info("name", "parent", "runtime.host", "resourcePool", "datastore", "layoutEx.disk")
	if err != nil {
		t.Fatalf("Cannot read VM properties: %v", err)
	}

	if vmInfo.Name != vmName {
		t.Errorf("Invalid VM name: expected '%v', got '%v'", vmName, vmInfo.Name)
	}

	f := d.NewFolder(vmInfo.Parent)
	folderPath, err := f.Path()
	if err != nil {
		t.Fatalf("Cannot read folder name: %v", err)
	}
	if folderPath != "" {
		t.Errorf("Invalid folder: expected '/', got '%v'", folderPath)
	}

	h := d.NewHost(vmInfo.Runtime.Host)
	hostInfo, err := h.Info("name")
	if err != nil {
		t.Fatal("Cannot read host properties: ", err)
	}
	if hostInfo.Name != hostName {
		t.Errorf("Invalid host name: expected '%v', got '%v'", hostName, hostInfo.Name)
	}

	p := d.NewResourcePool(vmInfo.ResourcePool)
	poolPath, err := p.Path()
	if err != nil {
		t.Fatalf("Cannot read resource pool name: %v", err)
	}
	if poolPath != "" {
		t.Error("Invalid resource pool: expected '/', got '%v'", poolPath)
	}

	dsr := vmInfo.Datastore[0].Reference()
	ds := d.NewDatastore(&dsr)
	dsInfo, err := ds.Info("name")
	if err != nil {
		t.Fatal("Cannot read datastore properties: ", err)
	}
	if dsInfo.Name != "datastore1" {
		t.Errorf("Invalid datastore name: expected '%v', got '%v'", "datastore1", dsInfo.Name)
	}

	if len(vmInfo.LayoutEx.Disk[0].Chain) != 1 {
		t.Error("Not a full clone")
	}
}

func TestVMAcc_folder(t *testing.T) {

}

func TestVMAcc_hardware(t *testing.T) {
	_ /*d*/, vm, _ /*vmName*/, vmDestructor := initVMAccTest(t)
	defer vmDestructor()

	log.Printf("[DEBUG] Configuring the vm")
	config := &HardwareConfig{
		CPUs:           2,
		CPUReservation: 1000,
		CPULimit:       1500,
		RAM:            2048,
		RAMReservation: 1024,
	}
	vm.Configure(config)

	log.Printf("[DEBUG] Running checks")
	vmInfo, err := vm.Info("config")
	if err != nil {
		t.Fatalf("Cannot read VM properties: %v", err)
	}

	cpuSockets := vmInfo.Config.Hardware.NumCPU
	if cpuSockets != config.CPUs {
		t.Errorf("VM should have %v CPU sockets, got %v", config.CPUs, cpuSockets)
	}

	cpuReservation := vmInfo.Config.CpuAllocation.GetResourceAllocationInfo().Reservation
	if cpuReservation != config.CPUReservation {
		t.Errorf("VM should have CPU reservation for %v Mhz, got %v", config.CPUReservation, cpuReservation)
	}

	cpuLimit := vmInfo.Config.CpuAllocation.GetResourceAllocationInfo().Limit
	if cpuLimit != config.CPULimit {
		t.Errorf("VM should have CPU reservation for %v Mhz, got %v", config.CPULimit, cpuLimit)
	}

	ram := vmInfo.Config.Hardware.MemoryMB
	if int64(ram) != config.RAM {
		t.Errorf("VM should have %v MB of RAM, got %v", config.RAM, ram)
	}

	ramReservation := vmInfo.Config.MemoryAllocation.GetResourceAllocationInfo().Reservation
	if ramReservation != config.RAMReservation {
		t.Errorf("VM should have RAM reservation for %v MB, got %v", config.RAMReservation, ramReservation)
	}
}

func startVM(t *testing.T, vm *VirtualMachine, vmName string) (stopper func()) {
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
	vm.WaitForShutdown(1 * time.Minute)
}

func TestVMAcc_snapshot(t *testing.T) {
	_ /*d*/, vm, vmName, vmDestructor := initVMAccTest(t)
	defer vmDestructor()

	stopper := startVM(t, vm, vmName)
	defer stopper()

	vm.CreateSnapshot("test-snapshot")

	vmInfo, err := vm.Info("layoutEx.disk")
	if err != nil {
		t.Fatalf("Cannot read VM properties: %v", err)
	}

	layers := len(vmInfo.LayoutEx.Disk[0].Chain)
	if layers != 2 {
		t.Errorf("VM should have a single snapshot. expected 2 disk layers, got %v", layers)
	}
}

func TestVMAcc_template(t *testing.T) {
	_ /*d*/, vm, _ /*vmName*/, vmDestructor := initVMAccTest(t)
	defer vmDestructor()

	vm.ConvertToTemplate()
	vmInfo, err := vm.Info("config.template")
	if err != nil {
		t.Errorf("Cannot read VM properties: %v", err)
	} else if !vmInfo.Config.Template {
		t.Error("Not a template")
	}
}

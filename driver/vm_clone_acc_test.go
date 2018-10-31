package driver

import (
	"context"
	"log"
	"net"
	"testing"
	"time"
)

func TestVMAcc_clone(t *testing.T) {
	testCases := []struct {
		name          string
		config        *CloneConfig
		checkFunction func(*testing.T, *VirtualMachine, *CloneConfig)
	}{
		{"Default", &CloneConfig{}, cloneDefaultCheck},
		{"LinkedClone", &CloneConfig{LinkedClone: true}, cloneLinkedCloneCheck},
		{"Folder", &CloneConfig{LinkedClone: true, Folder: "folder1/folder2"}, cloneFolderCheck},
		{"ResourcePool", &CloneConfig{LinkedClone: true, ResourcePool: "pool1/pool2"}, cloneResourcePoolCheck},
		{"Configure", &CloneConfig{LinkedClone: true}, configureCheck},
		{"Configure_RAMReserveAll", &CloneConfig{LinkedClone: true}, configureRAMReserveAllCheck},
		{"StartAndStop", &CloneConfig{LinkedClone: true}, startAndStopCheck},
		{"Template", &CloneConfig{LinkedClone: true}, templateCheck},
		{"Snapshot", &CloneConfig{}, snapshotCheck},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.config.Host = TestHostName
			tc.config.Name = newVMName()

			templateName := "alpine"
			d := newTestDriver(t)

			template, err := d.FindVM(templateName) // Don't destroy this VM!
			if err != nil {
				t.Fatalf("Cannot find template vm '%v': %v", templateName, err)
			}

			log.Printf("[DEBUG] Clonning VM")
			vm, err := template.Clone(context.TODO(), tc.config)
			if err != nil {
				t.Fatalf("Cannot clone vm '%v': %v", templateName, err)
			}

			defer destroyVM(t, vm, tc.config.Name)

			log.Printf("[DEBUG] Running check function")
			tc.checkFunction(t, vm, tc.config)
		})
	}
}

func cloneDefaultCheck(t *testing.T, vm *VirtualMachine, config *CloneConfig) {
	d := vm.driver

	// Check that the clone can be found by its name
	if _, err := d.FindVM(config.Name); err != nil {
		t.Errorf("Cannot find created vm '%v': %v", config.Name, err)
	}

	vmInfo, err := vm.Info("name", "parent", "runtime.host", "resourcePool", "datastore", "layoutEx.disk")
	if err != nil {
		t.Fatalf("Cannot read VM properties: %v", err)
	}

	if vmInfo.Name != config.Name {
		t.Errorf("Invalid VM name: expected '%v', got '%v'", config.Name, vmInfo.Name)
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
	if hostInfo.Name != TestHostName {
		t.Errorf("Invalid host name: expected '%v', got '%v'", TestHostName, hostInfo.Name)
	}

	p := d.NewResourcePool(vmInfo.ResourcePool)
	poolPath, err := p.Path()
	if err != nil {
		t.Fatalf("Cannot read resource pool name: %v", err)
	}
	if poolPath != "" {
		t.Errorf("Invalid resource pool: expected '/', got '%v'", poolPath)
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

func configureCheck(t *testing.T, vm *VirtualMachine, _ *CloneConfig) {
	log.Printf("[DEBUG] Configuring the vm")
	hwConfig := &HardwareConfig{
		CPUs:                2,
		CPUReservation:      1000,
		CPULimit:            1500,
		RAM:                 2048,
		RAMReservation:      1024,
		MemoryHotAddEnabled: true,
		CpuHotAddEnabled:    true,
	}
	vm.Configure(hwConfig)

	log.Printf("[DEBUG] Running checks")
	vmInfo, err := vm.Info("config")
	if err != nil {
		t.Fatalf("Cannot read VM properties: %v", err)
	}

	cpuSockets := vmInfo.Config.Hardware.NumCPU
	if cpuSockets != hwConfig.CPUs {
		t.Errorf("VM should have %v CPU sockets, got %v", hwConfig.CPUs, cpuSockets)
	}

	cpuReservation := *vmInfo.Config.CpuAllocation.Reservation
	if cpuReservation != hwConfig.CPUReservation {
		t.Errorf("VM should have CPU reservation for %v Mhz, got %v", hwConfig.CPUReservation, cpuReservation)
	}

	cpuLimit := *vmInfo.Config.CpuAllocation.Limit
	if cpuLimit != hwConfig.CPULimit {
		t.Errorf("VM should have CPU reservation for %v Mhz, got %v", hwConfig.CPULimit, cpuLimit)
	}

	ram := vmInfo.Config.Hardware.MemoryMB
	if int64(ram) != hwConfig.RAM {
		t.Errorf("VM should have %v MB of RAM, got %v", hwConfig.RAM, ram)
	}

	ramReservation := *vmInfo.Config.MemoryAllocation.Reservation
	if ramReservation != hwConfig.RAMReservation {
		t.Errorf("VM should have RAM reservation for %v MB, got %v", hwConfig.RAMReservation, ramReservation)
	}

	cpuHotAdd := vmInfo.Config.CpuHotAddEnabled
	if *cpuHotAdd != hwConfig.CpuHotAddEnabled {
		t.Errorf("VM should have CPU hot add set to %v, got %v", hwConfig.CpuHotAddEnabled, cpuHotAdd)
	}

	memoryHotAdd := vmInfo.Config.MemoryHotAddEnabled
	if *memoryHotAdd != hwConfig.MemoryHotAddEnabled {
		t.Errorf("VM should have Memroy hot add set to %v, got %v", hwConfig.MemoryHotAddEnabled, memoryHotAdd)
	}
}

func configureRAMReserveAllCheck(t *testing.T, vm *VirtualMachine, _ *CloneConfig) {
	log.Printf("[DEBUG] Configuring the vm")
	vm.Configure(&HardwareConfig{RAMReserveAll: true})

	log.Printf("[DEBUG] Running checks")
	vmInfo, err := vm.Info("config")
	if err != nil {
		t.Fatalf("Cannot read VM properties: %v", err)
	}

	if *vmInfo.Config.MemoryReservationLockedToMax != true {
		t.Errorf("VM should have all RAM reserved")
	}
}

func cloneLinkedCloneCheck(t *testing.T, vm *VirtualMachine, _ *CloneConfig) {
	vmInfo, err := vm.Info("layoutEx.disk")
	if err != nil {
		t.Fatalf("Cannot read VM properties: %v", err)
	}

	if len(vmInfo.LayoutEx.Disk[0].Chain) != 2 {
		t.Error("Not a linked clone")
	}
}

func cloneFolderCheck(t *testing.T, vm *VirtualMachine, config *CloneConfig) {
	vmInfo, err := vm.Info("parent")
	if err != nil {
		t.Fatalf("Cannot read VM properties: %v", err)
	}

	f := vm.driver.NewFolder(vmInfo.Parent)
	path, err := f.Path()
	if err != nil {
		t.Fatalf("Cannot read folder name: %v", err)
	}
	if path != config.Folder {
		t.Errorf("Wrong folder. expected: %v, got: %v", config.Folder, path)
	}
}

func cloneResourcePoolCheck(t *testing.T, vm *VirtualMachine, config *CloneConfig) {
	vmInfo, err := vm.Info("resourcePool")
	if err != nil {
		t.Fatalf("Cannot read VM properties: %v", err)
	}

	p := vm.driver.NewResourcePool(vmInfo.ResourcePool)
	path, err := p.Path()
	if err != nil {
		t.Fatalf("Cannot read resource pool name: %v", err)
	}
	if path != config.ResourcePool {
		t.Errorf("Wrong folder. expected: %v, got: %v", config.ResourcePool, path)
	}
}

func startAndStopCheck(t *testing.T, vm *VirtualMachine, config *CloneConfig) {
	stopper := startVM(t, vm, config.Name)
	defer stopper()

	switch ip, err := vm.WaitForIP(context.TODO()); {
	case err != nil:
		t.Errorf("Cannot obtain IP address from created vm '%v': %v", config.Name, err)
	case net.ParseIP(ip) == nil:
		t.Errorf("'%v' is not a valid ip address", ip)
	}

	vm.StartShutdown()
	log.Printf("[DEBUG] Waiting max 1m0s for shutdown to complete")
	vm.WaitForShutdown(context.TODO(), 1*time.Minute)
}

func snapshotCheck(t *testing.T, vm *VirtualMachine, config *CloneConfig) {
	stopper := startVM(t, vm, config.Name)
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

func templateCheck(t *testing.T, vm *VirtualMachine, _ *CloneConfig) {
	vm.ConvertToTemplate()
	vmInfo, err := vm.Info("config.template")
	if err != nil {
		t.Errorf("Cannot read VM properties: %v", err)
	} else if !vmInfo.Config.Template {
		t.Error("Not a template")
	}
}

func startVM(t *testing.T, vm *VirtualMachine, vmName string) (stopper func()) {
	log.Printf("[DEBUG] Starting the vm")
	if err := vm.PowerOn(); err != nil {
		t.Fatalf("Cannot start vm '%v': %v", vmName, err)
	}
	return func() {
		log.Printf("[DEBUG] Powering off the vm")
		if err := vm.PowerOff(); err != nil {
			t.Errorf("Cannot power off started vm '%v': %v", vmName, err)
		}
	}
}

func destroyVM(t *testing.T, vm *VirtualMachine, vmName string) {
	log.Printf("[DEBUG] Deleting the VM")
	if err := vm.Destroy(); err != nil {
		t.Errorf("!!! ERROR DELETING VM '%v': %v!!!", vmName, err)
	}

	// Check that the clone is no longer exists
	if _, err := vm.driver.FindVM(vmName); err == nil {
		t.Errorf("!!! STILL CAN FIND VM '%v'. IT MIGHT NOT HAVE BEEN DELETED !!!", vmName)
	}
}

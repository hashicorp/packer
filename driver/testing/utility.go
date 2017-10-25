package testing

import (
	"fmt"
	"github.com/jetbrains-infra/packer-builder-vsphere/driver"
	"math/rand"
	"os"
	"testing"
	"time"
)

func NewTestDriver(t *testing.T) *driver.Driver {
	d, err := driver.NewDriver(&driver.ConnectConfig{
		VCenterServer:      DefaultVCenterServer,
		Username:           DefaultVCenterUsername,
		Password:           DefaultVCenterPassword,
		InsecureConnection: true,
	})
	if err != nil {
		t.Fatal("Cannot connect: ", err)
	}
	return d
}

func NewVMName() string {
	rand.Seed(time.Now().UTC().UnixNano())
	return fmt.Sprintf("test-%v", rand.Intn(1000))
}

func CheckDatastoreName(t *testing.T, ds *driver.Datastore, name string) {
	info, err := ds.Info("name")
	if err != nil {
		t.Fatalf("Cannot read datastore properties: %v", err)
	}
	if info.Name != name {
		t.Errorf("Wrong datastore. expected: %v, got: %v", name, info.Name)
	}
}

func initDriverAcceptanceTest(t *testing.T) {
	// We only run acceptance tests if an env var is set because they're
	// slow and require outside configuration.
	if os.Getenv(TestEnvVar) == "" {
		t.Skip(fmt.Sprintf(
			"Acceptance tests skipped unless env '%s' set",
			TestEnvVar))
	}

	// We require verbose mode so that the user knows what is going on.
	if !testing.Verbose() {
		t.Fatal("Acceptance tests must be run with the -v flag on tests")
	}
}

func VMCheckDefault(t *testing.T, d *driver.Driver, vm *driver.VirtualMachine,
					name string, host string, datastore string) error {
	vmInfo, err := vm.Info("name", "parent", "runtime.host", "resourcePool", "datastore", "layoutEx.disk")
	if err != nil {
		t.Fatalf("Cannot read VM properties: %v", err)
	}

	if vmInfo.Name != name {
		t.Errorf("Invalid VM name: expected '%v', got '%v'", name, vmInfo.Name)
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
	if hostInfo.Name != host {
		t.Errorf("Invalid host name: expected '%v', got '%v'", host, hostInfo.Name)
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
	if dsInfo.Name != datastore {
		t.Errorf("Invalid datastore name: expected '%v', got '%v'", datastore, dsInfo.Name)
	}

	if len(vmInfo.LayoutEx.Disk[0].Chain) != 1 {
		t.Error("Not a full clone")
	}

	return nil
}

func VMCheckHardware(t* testing.T, d *driver.Driver, vm *driver.VirtualMachine) error {
	vmInfo, err := vm.Info("config")
	if err != nil {
		t.Fatalf("Cannot read VM properties: %v", err)
	}

	cpuSockets := vmInfo.Config.Hardware.NumCPU
	if cpuSockets != 2 {
		t.Errorf("VM should have 2 CPU sockets, got %v", cpuSockets)
	}

	cpuReservation := vmInfo.Config.CpuAllocation.GetResourceAllocationInfo().Reservation
	if cpuReservation != 1000 {
		t.Errorf("VM should have CPU reservation for 1000 Mhz, got %v", cpuReservation)
	}

	cpuLimit := vmInfo.Config.CpuAllocation.GetResourceAllocationInfo().Limit
	if cpuLimit != 1500 {
		t.Errorf("VM should have CPU reservation for 1500 Mhz, got %v", cpuLimit)
	}

	ram := vmInfo.Config.Hardware.MemoryMB
	if ram != 2048 {
		t.Errorf("VM should have 2048 MB of RAM, got %v", ram)
	}

	ramReservation := vmInfo.Config.MemoryAllocation.GetResourceAllocationInfo().Reservation
	if ramReservation != 1024 {
		t.Errorf("VM should have RAM reservation for 1024 MB, got %v", ramReservation)
	}

	return nil
}

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
		t.Fatalf("Cannot connect: %v", err)
	}
	return d
}

func NewVMName() string {
	rand.Seed(time.Now().UTC().UnixNano())
	return fmt.Sprintf("test-%v", rand.Intn(1000))
}

func CheckDatastoreName(t *testing.T, ds *driver.Datastore, name string) {
	switch info, err := ds.Info("name"); {
	case err != nil:
		t.Errorf("Cannot read datastore properties: %v", err)
	case info.Name != name:
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
					name string, host string, datastore string) {
	vmInfo, err := vm.Info("name", "parent", "runtime.host", "resourcePool", "datastore", "layoutEx.disk")
	if err != nil {
		t.Fatalf("Cannot read VM properties: %v", err)
	}

	if vmInfo.Name != name {
		t.Errorf("Invalid VM name: expected '%v', got '%v'", name, vmInfo.Name)
	}

	f := d.NewFolder(vmInfo.Parent)
	switch folderPath, err := f.Path(); {
	case err != nil:
		t.Errorf("Cannot read folder name: %v", err)
	case folderPath != "":
		t.Errorf("Invalid folder: expected '/', got '%v'", folderPath)
	}

	h := d.NewHost(vmInfo.Runtime.Host)
	switch hostInfo, err := h.Info("name"); {
	case err != nil:
		t.Errorf("Cannot read host properties: %v", err)
	case hostInfo.Name != host:
		t.Errorf("Invalid host name: expected '%v', got '%v'", host, hostInfo.Name)
	}

	p := d.NewResourcePool(vmInfo.ResourcePool)
	switch poolPath, err := p.Path(); {
	case err != nil:
		t.Errorf("Cannot read resource pool name: %v", err)
	case poolPath != "":
		t.Error("Invalid resource pool: expected '/', got '%v'", poolPath)
	}


	dsr := vmInfo.Datastore[0].Reference()
	ds := d.NewDatastore(&dsr)
	switch dsInfo, err := ds.Info("name"); {
	case err != nil:
		t.Errorf("Cannot read datastore properties: %v", err)
	case dsInfo.Name != datastore:
		t.Errorf("Invalid datastore name: expected '%v', got '%v'", datastore, dsInfo.Name)
	}

	if len(vmInfo.LayoutEx.Disk[0].Chain) != 1 {
		t.Error("Not a full clone")
	}
}

func VMCheckHardware(t* testing.T, d *driver.Driver, vm *driver.VirtualMachine) {
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
}

func VMCheckTemplate(t* testing.T, d *driver.Driver, vm *driver.VirtualMachine) {
	switch vmInfo, err := vm.Info("config.template"); {
	case err != nil:
		t.Errorf("Cannot read VM properties: %v", err)
	case !vmInfo.Config.Template:
		t.Error("Not a template")
	}
}

func VMCheckDatastore(t* testing.T, d *driver.Driver, vm *driver.VirtualMachine, name string) {
	vmInfo, err := vm.Info("datastore")
	if err != nil {
		t.Fatalf("Cannot read VM properties: %v", err)
	}

	n := len(vmInfo.Datastore)
	if n != 1 {
		t.Fatalf("VM should have 1 datastore, got %v", n)
	}

	ds := d.NewDatastore(&vmInfo.Datastore[0])
	CheckDatastoreName(t, ds, name)
}

func VMCheckSnapshor(t* testing.T, d *driver.Driver, vm *driver.VirtualMachine) {
	vmInfo, err := vm.Info("layoutEx.disk")
	if err != nil {
		t.Fatalf("Cannot read VM properties: %v", err)
	}

	layers := len(vmInfo.LayoutEx.Disk[0].Chain)
	if layers != 2 {
		t.Errorf("VM should have a single snapshot. expected 2 disk layers, got %v", layers)
	}
}

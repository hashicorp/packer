package iso

import (
	builderT "github.com/hashicorp/packer/helper/builder/testing"
	commonT "github.com/jetbrains-infra/packer-builder-vsphere/common/testing"
	"testing"
	"github.com/hashicorp/packer/packer"
	"github.com/vmware/govmomi/vim25/types"
	"fmt"
	"io/ioutil"
)

func TestISOBuilderAcc_default(t *testing.T) {
	config := defaultConfig()
	builderT.Test(t, builderT.TestCase{
		Builder:  &Builder{},
		Template: commonT.RenderConfig(config),
		Check:    checkDefault(t, config["vm_name"].(string), config["host"].(string), "datastore1"),
	})
}

func defaultConfig() map[string]interface{} {
	config := map[string]interface{}{
		"vcenter_server":      "vcenter.vsphere65.test",
		"username":            "root",
		"password":            "jetbrains",
		"insecure_connection": true,

		"host": "esxi-1.vsphere65.test",

		"ssh_username": "root",
		"ssh_password": "jetbrains",

		"vm_name":   commonT.NewVMName(),
		"disk_size": 2,

		"communicator": "none", // do not start the VM without any bootable devices
	}

	return config
}

func checkDefault(t *testing.T, name string, host string, datastore string) builderT.TestCheckFunc {
	return func(artifacts []packer.Artifact) error {
		d := commonT.TestConn(t)
		vm := commonT.GetVM(t, d, artifacts)

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
			t.Errorf("Invalid resource pool: expected '/', got '%v'", poolPath)
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

		return nil
	}
}

func TestISOBuilderAcc_hardware(t *testing.T) {
	builderT.Test(t, builderT.TestCase{
		Builder:  &Builder{},
		Template: hardwareConfig(),
		Check:    checkHardware(t),
	})
}

func hardwareConfig() string {
	config := defaultConfig()
	config["CPUs"] = 2
	config["CPU_reservation"] = 1000
	config["CPU_limit"] = 1500
	config["RAM"] = 2048
	config["RAM_reservation"] = 1024

	return commonT.RenderConfig(config)
}

func checkHardware(t *testing.T) builderT.TestCheckFunc {
	return func(artifacts []packer.Artifact) error {
		d := commonT.TestConn(t)

		vm := commonT.GetVM(t, d, artifacts)
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
}

func TestISOBuilderAcc_cdrom(t *testing.T) {
	builderT.Test(t, builderT.TestCase{
		Builder:  &Builder{},
		Template: cdromConfig(),
	})
}

func cdromConfig() string {
	config := defaultConfig()
	config["iso_paths"] = []string{
		"[datastore1] test0.iso",
		"[datastore1] test1.iso",
	}
	return commonT.RenderConfig(config)
}

func TestISOBuilderAcc_networkCard(t *testing.T) {
	builderT.Test(t, builderT.TestCase{
		Builder:  &Builder{},
		Template: networkCardConfig(),
		Check:    checkNetworkCard(t),
	})
}

func networkCardConfig() string {
	config := defaultConfig()
	config["network_card"] = "vmxnet3"
	return commonT.RenderConfig(config)
}

func checkNetworkCard(t *testing.T) builderT.TestCheckFunc {
	return func(artifacts []packer.Artifact) error {
		d := commonT.TestConn(t)

		vm := commonT.GetVM(t, d, artifacts)
		devices, err := vm.Devices()
		if err != nil {
			t.Fatalf("Cannot read VM properties: %v", err)
		}

		netCards := devices.SelectByType((*types.VirtualEthernetCard)(nil))
		if len(netCards) == 0 {
			t.Fatalf("Cannot find the network card")
		}
		if len(netCards) > 1 {
			t.Fatalf("Found several network catds")
		}
		if _, ok := netCards[0].(*types.VirtualVmxnet3); !ok {
			t.Errorf("The network card type is not the expected one (vmxnet3)")
		}

		return nil
	}
}

func TestISOBuilderAcc_createFloppy(t *testing.T) {
	tmpFile, err := ioutil.TempFile("", "packer-vsphere-iso-test")
	if err != nil {
		t.Fatalf("Error creating temp file ")
	}
	fmt.Fprint(tmpFile, "Hello, World!")
	tmpFile.Close()

	builderT.Test(t, builderT.TestCase{
		Builder:  &Builder{},
		Template: createFloppyConfig(tmpFile.Name()),
	})
}

func createFloppyConfig(filePath string) string {
	config := defaultConfig()
	config["floppy_files"] = []string{filePath}
	return commonT.RenderConfig(config)
}

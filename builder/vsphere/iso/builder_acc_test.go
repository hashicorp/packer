package iso

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	commonT "github.com/hashicorp/packer/builder/vsphere/common/testing"
	builderT "github.com/hashicorp/packer/helper/builder/testing"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/vmware/govmomi/vim25/types"
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
	username := os.Getenv("VSPHERE_USERNAME")
	if username == "" {
		username = "root"
	}
	password := os.Getenv("VSPHERE_PASSWORD")
	if password == "" {
		password = "jetbrains"
	}

	config := map[string]interface{}{
		"vcenter_server":      "vcenter.vsphere65.test",
		"username":            username,
		"password":            password,
		"insecure_connection": true,

		"host": "esxi-1.vsphere65.test",

		"ssh_username": "root",
		"ssh_password": "jetbrains",

		"vm_name": commonT.NewVMName(),
		"storage": map[string]interface{}{
			"disk_size": 2048,
		},

		"communicator": "none", // do not start the VM without any bootable devices
	}

	return config
}

func checkDefault(t *testing.T, name string, host string, datastore string) builderT.TestCheckFunc {
	return func(artifacts []packersdk.Artifact) error {
		d := commonT.TestConn(t)
		vm := commonT.GetVM(t, d, artifacts)

		vmInfo, err := vm.Info("name", "parent", "runtime.host", "resourcePool", "datastore", "layoutEx.disk", "config.firmware")
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

		fw := vmInfo.Config.Firmware
		if fw != "bios" {
			t.Errorf("Invalid firmware: expected 'bios', got '%v'", fw)
		}

		return nil
	}
}

func TestISOBuilderAcc_notes(t *testing.T) {
	builderT.Test(t, builderT.TestCase{
		Builder:  &Builder{},
		Template: notesConfig(),
		Check:    checkNotes(t),
	})
}

func notesConfig() string {
	config := defaultConfig()
	config["notes"] = "test"

	return commonT.RenderConfig(config)
}

func checkNotes(t *testing.T) builderT.TestCheckFunc {
	return func(artifacts []packersdk.Artifact) error {
		d := commonT.TestConn(t)
		vm := commonT.GetVM(t, d, artifacts)

		vmInfo, err := vm.Info("config.annotation")
		if err != nil {
			t.Fatalf("Cannot read VM properties: %v", err)
		}

		notes := vmInfo.Config.Annotation
		if notes != "test" {
			t.Errorf("notes should be 'test'")
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
	config["cpu_cores"] = 2
	config["CPU_reservation"] = 1000
	config["CPU_limit"] = 1500
	config["RAM"] = 2048
	config["RAM_reservation"] = 1024
	config["NestedHV"] = true
	config["firmware"] = "efi"
	config["video_ram"] = 8192

	return commonT.RenderConfig(config)
}

func checkHardware(t *testing.T) builderT.TestCheckFunc {
	return func(artifacts []packersdk.Artifact) error {
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

		cpuCores := vmInfo.Config.Hardware.NumCoresPerSocket
		if cpuCores != 2 {
			t.Errorf("VM should have 2 CPU cores per socket, got %v", cpuCores)
		}

		cpuReservation := *vmInfo.Config.CpuAllocation.Reservation
		if cpuReservation != 1000 {
			t.Errorf("VM should have CPU reservation for 1000 Mhz, got %v", cpuReservation)
		}

		cpuLimit := *vmInfo.Config.CpuAllocation.Limit
		if cpuLimit != 1500 {
			t.Errorf("VM should have CPU reservation for 1500 Mhz, got %v", cpuLimit)
		}

		ram := vmInfo.Config.Hardware.MemoryMB
		if ram != 2048 {
			t.Errorf("VM should have 2048 MB of RAM, got %v", ram)
		}

		ramReservation := *vmInfo.Config.MemoryAllocation.Reservation
		if ramReservation != 1024 {
			t.Errorf("VM should have RAM reservation for 1024 MB, got %v", ramReservation)
		}

		nestedHV := vmInfo.Config.NestedHVEnabled
		if !*nestedHV {
			t.Errorf("VM should have NestedHV enabled, got %v", nestedHV)
		}

		fw := vmInfo.Config.Firmware
		if fw != "efi" {
			t.Errorf("Invalid firmware: expected 'efi', got '%v'", fw)
		}

		l, err := vm.Devices()
		if err != nil {
			t.Fatalf("Cannot read VM devices: %v", err)
		}
		c := l.PickController((*types.VirtualIDEController)(nil))
		if c == nil {
			t.Errorf("VM should have IDE controller")
		}
		s := l.PickController((*types.VirtualAHCIController)(nil))
		if s != nil {
			t.Errorf("VM should have no SATA controllers")
		}

		v := l.SelectByType((*types.VirtualMachineVideoCard)(nil))
		if len(v) != 1 {
			t.Errorf("VM should have one video card")
		}
		if v[0].(*types.VirtualMachineVideoCard).VideoRamSizeInKB != 8192 {
			t.Errorf("Video RAM should be equal 8192")
		}

		return nil
	}
}

func TestISOBuilderAcc_limit(t *testing.T) {
	builderT.Test(t, builderT.TestCase{
		Builder:  &Builder{},
		Template: limitConfig(),
		Check:    checkLimit(t),
	})
}

func limitConfig() string {
	config := defaultConfig()
	config["CPUs"] = 1 // hardware is customized, but CPU limit is not specified explicitly

	return commonT.RenderConfig(config)
}

func checkLimit(t *testing.T) builderT.TestCheckFunc {
	return func(artifacts []packersdk.Artifact) error {
		d := commonT.TestConn(t)

		vm := commonT.GetVM(t, d, artifacts)
		vmInfo, err := vm.Info("config.cpuAllocation")
		if err != nil {
			t.Fatalf("Cannot read VM properties: %v", err)
		}

		limit := *vmInfo.Config.CpuAllocation.Limit
		if limit != -1 { // must be unlimited
			t.Errorf("Invalid CPU limit: expected '%v', got '%v'", -1, limit)
		}

		return nil
	}
}

func TestISOBuilderAcc_sata(t *testing.T) {
	builderT.Test(t, builderT.TestCase{
		Builder:  &Builder{},
		Template: sataConfig(),
		Check:    checkSata(t),
	})
}

func sataConfig() string {
	config := defaultConfig()
	config["cdrom_type"] = "sata"

	return commonT.RenderConfig(config)
}

func checkSata(t *testing.T) builderT.TestCheckFunc {
	return func(artifacts []packersdk.Artifact) error {
		d := commonT.TestConn(t)

		vm := commonT.GetVM(t, d, artifacts)

		l, err := vm.Devices()
		if err != nil {
			t.Fatalf("Cannot read VM devices: %v", err)
		}

		c := l.PickController((*types.VirtualAHCIController)(nil))
		if c == nil {
			t.Errorf("VM has no SATA controllers")
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
	config["network_adapters"] = map[string]interface{}{
		"network_card": "vmxnet3",
	}
	return commonT.RenderConfig(config)
}

func checkNetworkCard(t *testing.T) builderT.TestCheckFunc {
	return func(artifacts []packersdk.Artifact) error {
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
		t.Fatalf("Error creating temp file: %v", err)
	}
	_, err = fmt.Fprint(tmpFile, "Hello, World!")
	if err != nil {
		t.Fatalf("Error creating temp file: %v", err)
	}
	err = tmpFile.Close()
	if err != nil {
		t.Fatalf("Error creating temp file: %v", err)
	}

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

func TestISOBuilderAcc_full(t *testing.T) {
	config := fullConfig()
	builderT.Test(t, builderT.TestCase{
		Builder:  &Builder{},
		Template: commonT.RenderConfig(config),
		Check:    checkFull(t),
	})
}

func fullConfig() map[string]interface{} {
	username := os.Getenv("VSPHERE_USERNAME")
	if username == "" {
		username = "root"
	}
	password := os.Getenv("VSPHERE_PASSWORD")
	if password == "" {
		password = "jetbrains"
	}

	config := map[string]interface{}{
		"vcenter_server":      "vcenter.vsphere65.test",
		"username":            username,
		"password":            password,
		"insecure_connection": true,

		"vm_name": commonT.NewVMName(),
		"host":    "esxi-1.vsphere65.test",

		"RAM": 512,
		"disk_controller_type": []string{
			"pvscsi",
		},
		"storage": map[string]interface{}{
			"disk_size":             1024,
			"disk_thin_provisioned": true,
		},
		"network_adapters": map[string]interface{}{
			"network_card": "vmxnet3",
		},
		"guest_os_type": "other3xLinux64Guest",

		"iso_paths": []string{
			"[datastore1] ISO/alpine-standard-3.8.2-x86_64.iso",
		},
		"floppy_files": []string{
			"../examples/alpine/answerfile",
			"../examples/alpine/setup.sh",
		},

		"boot_wait": "20s",
		"boot_command": []string{
			"root<enter><wait>",
			"mount -t vfat /dev/fd0 /media/floppy<enter><wait>",
			"setup-alpine -f /media/floppy/answerfile<enter>",
			"<wait5>",
			"jetbrains<enter>",
			"jetbrains<enter>",
			"<wait5>",
			"y<enter>",
			"<wait10><wait10><wait10><wait10>",
			"reboot<enter>",
			"<wait10><wait10><wait10>",
			"root<enter>",
			"jetbrains<enter><wait>",
			"mount -t vfat /dev/fd0 /media/floppy<enter><wait>",
			"/media/floppy/SETUP.SH<enter>",
		},

		"ssh_username": "root",
		"ssh_password": "jetbrains",
	}

	return config
}

func checkFull(t *testing.T) builderT.TestCheckFunc {
	return func(artifacts []packersdk.Artifact) error {
		d := commonT.TestConn(t)
		vm := commonT.GetVM(t, d, artifacts)

		vmInfo, err := vm.Info("config.bootOptions")
		if err != nil {
			t.Fatalf("Cannot read VM properties: %v", err)
		}

		order := vmInfo.Config.BootOptions.BootOrder
		if order != nil {
			t.Errorf("Boot order must be empty")
		}

		devices, err := vm.Devices()
		if err != nil {
			t.Fatalf("Cannot read devices: %v", err)
		}
		cdroms := devices.SelectByType((*types.VirtualCdrom)(nil))
		for _, cd := range cdroms {
			_, ok := cd.(*types.VirtualCdrom).Backing.(*types.VirtualCdromRemotePassthroughBackingInfo)
			if !ok {
				t.Errorf("wrong cdrom backing")
			}
		}

		return nil
	}
}

func TestISOBuilderAcc_bootOrder(t *testing.T) {
	config := fullConfig()
	config["boot_order"] = "disk,cdrom,floppy"

	builderT.Test(t, builderT.TestCase{
		Builder:  &Builder{},
		Template: commonT.RenderConfig(config),
		Check:    checkBootOrder(t),
	})
}

func checkBootOrder(t *testing.T) builderT.TestCheckFunc {
	return func(artifacts []packersdk.Artifact) error {
		d := commonT.TestConn(t)
		vm := commonT.GetVM(t, d, artifacts)

		vmInfo, err := vm.Info("config.bootOptions")
		if err != nil {
			t.Fatalf("Cannot read VM properties: %v", err)
		}

		order := vmInfo.Config.BootOptions.BootOrder
		if order == nil {
			t.Errorf("Boot order must not be empty")
		}

		return nil
	}
}

func TestISOBuilderAcc_cluster(t *testing.T) {
	builderT.Test(t, builderT.TestCase{
		Builder:  &Builder{},
		Template: clusterConfig(),
	})
}

func clusterConfig() string {
	config := defaultConfig()
	config["cluster"] = "cluster1"
	config["host"] = "esxi-2.vsphere65.test"

	return commonT.RenderConfig(config)
}

func TestISOBuilderAcc_clusterDRS(t *testing.T) {
	builderT.Test(t, builderT.TestCase{
		Builder:  &Builder{},
		Template: clusterDRSConfig(),
	})
}

func clusterDRSConfig() string {
	config := defaultConfig()
	config["cluster"] = "cluster2"
	config["host"] = ""
	config["datastore"] = "datastore3" // bug #183
	config["network_adapters"] = map[string]interface{}{
		"network": "VM Network",
	}

	return commonT.RenderConfig(config)
}

package clone

import (
	"os"
	"testing"

	"github.com/hashicorp/packer/builder/vsphere/common"
	commonT "github.com/hashicorp/packer/builder/vsphere/common/testing"
	builderT "github.com/hashicorp/packer/helper/builder/testing"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/vmware/govmomi/vim25/types"
)

func TestCloneBuilderAcc_default(t *testing.T) {
	t.Skip("Acceptance tests not configured yet.")
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

		"template": "alpine",
		"host":     "esxi-1.vsphere65.test",

		"linked_clone": true, // speed up
		"communicator": "none",
	}
	config["vm_name"] = commonT.NewVMName()
	return config
}

func checkDefault(t *testing.T, name string, host string, datastore string) builderT.TestCheckFunc {
	return func(artifacts []packersdk.Artifact) error {
		d := commonT.TestConn(t)
		vm := commonT.GetVM(t, d, artifacts)

		vmInfo, err := vm.Info("name", "parent", "runtime.host", "resourcePool", "datastore")
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

func TestCloneBuilderAcc_artifact(t *testing.T) {
	t.Skip("Acceptance tests not configured yet.")
	config := defaultConfig()
	builderT.Test(t, builderT.TestCase{
		Builder:  &Builder{},
		Template: commonT.RenderConfig(config),
		Check:    checkArtifact(t),
	})
}

func checkArtifact(t *testing.T) builderT.TestCheckFunc {
	return func(artifacts []packersdk.Artifact) error {
		if len(artifacts) > 1 {
			t.Fatal("more than 1 artifact")
		}

		artifactRaw := artifacts[0]
		_, ok := artifactRaw.(*common.Artifact)
		if !ok {
			t.Fatalf("unknown artifact: %#v", artifactRaw)
		}

		return nil
	}
}

func TestCloneBuilderAcc_folder(t *testing.T) {
	t.Skip("Acceptance tests not configured yet.")
	builderT.Test(t, builderT.TestCase{
		Builder:  &Builder{},
		Template: folderConfig(),
		Check:    checkFolder(t, "folder1/folder2"),
	})
}

func folderConfig() string {
	config := defaultConfig()
	config["folder"] = "folder1/folder2"
	return commonT.RenderConfig(config)
}

func checkFolder(t *testing.T, folder string) builderT.TestCheckFunc {
	return func(artifacts []packersdk.Artifact) error {
		d := commonT.TestConn(t)
		vm := commonT.GetVM(t, d, artifacts)

		vmInfo, err := vm.Info("parent")
		if err != nil {
			t.Fatalf("Cannot read VM properties: %v", err)
		}

		f := d.NewFolder(vmInfo.Parent)
		path, err := f.Path()
		if err != nil {
			t.Fatalf("Cannot read folder name: %v", err)
		}
		if path != folder {
			t.Errorf("Wrong folder. expected: %v, got: %v", folder, path)
		}

		return nil
	}
}

func TestCloneBuilderAcc_resourcePool(t *testing.T) {
	t.Skip("Acceptance tests not configured yet.")
	builderT.Test(t, builderT.TestCase{
		Builder:  &Builder{},
		Template: resourcePoolConfig(),
		Check:    checkResourcePool(t, "pool1/pool2"),
	})
}

func resourcePoolConfig() string {
	config := defaultConfig()
	config["resource_pool"] = "pool1/pool2"
	return commonT.RenderConfig(config)
}

func checkResourcePool(t *testing.T, pool string) builderT.TestCheckFunc {
	return func(artifacts []packersdk.Artifact) error {
		d := commonT.TestConn(t)
		vm := commonT.GetVM(t, d, artifacts)

		vmInfo, err := vm.Info("resourcePool")
		if err != nil {
			t.Fatalf("Cannot read VM properties: %v", err)
		}

		p := d.NewResourcePool(vmInfo.ResourcePool)
		path, err := p.Path()
		if err != nil {
			t.Fatalf("Cannot read resource pool name: %v", err)
		}
		if path != pool {
			t.Errorf("Wrong folder. expected: %v, got: %v", pool, path)
		}

		return nil
	}
}

func TestCloneBuilderAcc_datastore(t *testing.T) {
	t.Skip("Acceptance tests not configured yet.")
	builderT.Test(t, builderT.TestCase{
		Builder:  &Builder{},
		Template: datastoreConfig(),
		Check:    checkDatastore(t, "datastore1"), // on esxi-1.vsphere65.test
	})
}

func datastoreConfig() string {
	config := defaultConfig()
	config["template"] = "alpine-host4" // on esxi-4.vsphere65.test
	config["linked_clone"] = false
	return commonT.RenderConfig(config)
}

func checkDatastore(t *testing.T, name string) builderT.TestCheckFunc {
	return func(artifacts []packersdk.Artifact) error {
		d := commonT.TestConn(t)
		vm := commonT.GetVM(t, d, artifacts)

		vmInfo, err := vm.Info("datastore")
		if err != nil {
			t.Fatalf("Cannot read VM properties: %v", err)
		}

		n := len(vmInfo.Datastore)
		if n != 1 {
			t.Fatalf("VM should have 1 datastore, got %v", n)
		}

		ds := d.NewDatastore(&vmInfo.Datastore[0])
		info, err := ds.Info("name")
		if err != nil {
			t.Fatalf("Cannot read datastore properties: %v", err)
		}
		if info.Name != name {
			t.Errorf("Wrong datastore. expected: %v, got: %v", name, info.Name)
		}

		return nil
	}
}

func TestCloneBuilderAcc_multipleDatastores(t *testing.T) {
	t.Skip("test must fail")

	builderT.Test(t, builderT.TestCase{
		Builder:  &Builder{},
		Template: multipleDatastoresConfig(),
	})
}

func multipleDatastoresConfig() string {
	config := defaultConfig()
	config["host"] = "esxi-4.vsphere65.test" // host with 2 datastores
	config["linked_clone"] = false
	return commonT.RenderConfig(config)
}

func TestCloneBuilderAcc_fullClone(t *testing.T) {
	t.Skip("Acceptance tests not configured yet.")
	builderT.Test(t, builderT.TestCase{
		Builder:  &Builder{},
		Template: fullCloneConfig(),
		Check:    checkFullClone(t),
	})
}

func fullCloneConfig() string {
	config := defaultConfig()
	config["linked_clone"] = false
	return commonT.RenderConfig(config)
}

func checkFullClone(t *testing.T) builderT.TestCheckFunc {
	return func(artifacts []packersdk.Artifact) error {
		d := commonT.TestConn(t)
		vm := commonT.GetVM(t, d, artifacts)

		vmInfo, err := vm.Info("layoutEx.disk")
		if err != nil {
			t.Fatalf("Cannot read VM properties: %v", err)
		}

		if len(vmInfo.LayoutEx.Disk[0].Chain) != 1 {
			t.Error("Not a full clone")
		}

		return nil
	}
}

func TestCloneBuilderAcc_linkedClone(t *testing.T) {
	t.Skip("Acceptance tests not configured yet.")
	builderT.Test(t, builderT.TestCase{
		Builder:  &Builder{},
		Template: linkedCloneConfig(),
		Check:    checkLinkedClone(t),
	})
}

func linkedCloneConfig() string {
	config := defaultConfig()
	config["linked_clone"] = true
	return commonT.RenderConfig(config)
}

func checkLinkedClone(t *testing.T) builderT.TestCheckFunc {
	return func(artifacts []packersdk.Artifact) error {
		d := commonT.TestConn(t)
		vm := commonT.GetVM(t, d, artifacts)

		vmInfo, err := vm.Info("layoutEx.disk")
		if err != nil {
			t.Fatalf("Cannot read VM properties: %v", err)
		}

		if len(vmInfo.LayoutEx.Disk[0].Chain) != 2 {
			t.Error("Not a linked clone")
		}

		return nil
	}
}

func TestCloneBuilderAcc_network(t *testing.T) {
	t.Skip("Acceptance tests not configured yet.")
	builderT.Test(t, builderT.TestCase{
		Builder:  &Builder{},
		Template: networkConfig(),
		Check:    checkNetwork(t, "VM Network 2"),
	})
}

func networkConfig() string {
	config := defaultConfig()
	config["template"] = "alpine-host4"
	config["host"] = "esxi-4.vsphere65.test"
	config["datastore"] = "datastore4"
	config["network"] = "VM Network 2"
	return commonT.RenderConfig(config)
}

func checkNetwork(t *testing.T, name string) builderT.TestCheckFunc {
	return func(artifacts []packersdk.Artifact) error {
		d := commonT.TestConn(t)
		vm := commonT.GetVM(t, d, artifacts)

		vmInfo, err := vm.Info("network")
		if err != nil {
			t.Fatalf("Cannot read VM properties: %v", err)
		}

		n := len(vmInfo.Network)
		if n != 1 {
			t.Fatalf("VM should have 1 network, got %v", n)
		}

		ds := d.NewNetwork(&vmInfo.Network[0])
		info, err := ds.Info("name")
		if err != nil {
			t.Fatalf("Cannot read network properties: %v", err)
		}
		if info.Name != name {
			t.Errorf("Wrong network. expected: %v, got: %v", name, info.Name)
		}

		return nil
	}
}

func TestCloneBuilderAcc_hardware(t *testing.T) {
	t.Skip("Acceptance tests not configured yet.")
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
	config["CPU_hot_plug"] = true
	config["RAM_hot_plug"] = true
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

		cpuHotAdd := vmInfo.Config.CpuHotAddEnabled
		if !*cpuHotAdd {
			t.Errorf("VM should have CPU hot add enabled, got %v", cpuHotAdd)
		}

		memoryHotAdd := vmInfo.Config.MemoryHotAddEnabled
		if !*memoryHotAdd {
			t.Errorf("VM should have Memory hot add enabled, got %v", memoryHotAdd)
		}

		l, err := vm.Devices()
		if err != nil {
			t.Fatalf("Cannot read VM devices: %v", err)
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

func TestCloneBuilderAcc_RAMReservation(t *testing.T) {
	t.Skip("Acceptance tests not configured yet.")
	builderT.Test(t, builderT.TestCase{
		Builder:  &Builder{},
		Template: RAMReservationConfig(),
		Check:    checkRAMReservation(t),
	})
}

func RAMReservationConfig() string {
	config := defaultConfig()
	config["RAM_reserve_all"] = true

	return commonT.RenderConfig(config)
}

func checkRAMReservation(t *testing.T) builderT.TestCheckFunc {
	return func(artifacts []packersdk.Artifact) error {
		d := commonT.TestConn(t)

		vm := commonT.GetVM(t, d, artifacts)
		vmInfo, err := vm.Info("config")
		if err != nil {
			t.Fatalf("Cannot read VM properties: %v", err)
		}

		if *vmInfo.Config.MemoryReservationLockedToMax != true {
			t.Errorf("VM should have all RAM reserved")
		}

		return nil
	}
}

func TestCloneBuilderAcc_sshPassword(t *testing.T) {
	t.Skip("Acceptance tests not configured yet.")
	builderT.Test(t, builderT.TestCase{
		Builder:  &Builder{},
		Template: sshPasswordConfig(),
		Check:    checkDefaultBootOrder(t),
	})
}

func sshPasswordConfig() string {
	config := defaultConfig()
	config["communicator"] = "ssh"
	config["ssh_username"] = "root"
	config["ssh_password"] = "jetbrains"
	return commonT.RenderConfig(config)
}

func checkDefaultBootOrder(t *testing.T) builderT.TestCheckFunc {
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

		return nil
	}
}

func TestCloneBuilderAcc_sshKey(t *testing.T) {
	t.Skip("Acceptance tests not configured yet.")
	builderT.Test(t, builderT.TestCase{
		Builder:  &Builder{},
		Template: sshKeyConfig(),
	})
}

func sshKeyConfig() string {
	config := defaultConfig()
	config["communicator"] = "ssh"
	config["ssh_username"] = "root"
	config["ssh_private_key_file"] = "../test/test-key.pem"
	return commonT.RenderConfig(config)
}

func TestCloneBuilderAcc_snapshot(t *testing.T) {
	t.Skip("Acceptance tests not configured yet.")
	builderT.Test(t, builderT.TestCase{
		Builder:  &Builder{},
		Template: snapshotConfig(),
		Check:    checkSnapshot(t),
	})
}

func snapshotConfig() string {
	config := defaultConfig()
	config["linked_clone"] = false
	config["create_snapshot"] = true
	return commonT.RenderConfig(config)
}

func checkSnapshot(t *testing.T) builderT.TestCheckFunc {
	return func(artifacts []packersdk.Artifact) error {
		d := commonT.TestConn(t)

		vm := commonT.GetVM(t, d, artifacts)
		vmInfo, err := vm.Info("layoutEx.disk")
		if err != nil {
			t.Fatalf("Cannot read VM properties: %v", err)
		}

		layers := len(vmInfo.LayoutEx.Disk[0].Chain)
		if layers != 2 {
			t.Errorf("VM should have a single snapshot. expected 2 disk layers, got %v", layers)
		}

		return nil
	}
}

func TestCloneBuilderAcc_template(t *testing.T) {
	t.Skip("Acceptance tests not configured yet.")
	builderT.Test(t, builderT.TestCase{
		Builder:  &Builder{},
		Template: templateConfig(),
		Check:    checkTemplate(t),
	})
}

func templateConfig() string {
	config := defaultConfig()
	config["convert_to_template"] = true
	return commonT.RenderConfig(config)
}

func checkTemplate(t *testing.T) builderT.TestCheckFunc {
	return func(artifacts []packersdk.Artifact) error {
		d := commonT.TestConn(t)

		vm := commonT.GetVM(t, d, artifacts)
		vmInfo, err := vm.Info("config.template")
		if err != nil {
			t.Fatalf("Cannot read VM properties: %v", err)
		}

		if vmInfo.Config.Template != true {
			t.Error("Not a template")
		}

		return nil
	}
}

func TestCloneBuilderAcc_bootOrder(t *testing.T) {
	t.Skip("Acceptance tests not configured yet.")
	builderT.Test(t, builderT.TestCase{
		Builder:  &Builder{},
		Template: bootOrderConfig(),
		Check:    checkBootOrder(t),
	})
}

func bootOrderConfig() string {
	config := defaultConfig()
	config["communicator"] = "ssh"
	config["ssh_username"] = "root"
	config["ssh_password"] = "jetbrains"

	config["boot_order"] = "disk,cdrom,floppy"

	return commonT.RenderConfig(config)
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

func TestCloneBuilderAcc_notes(t *testing.T) {
	t.Skip("Acceptance tests not configured yet.")
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
			t.Errorf("notest should be 'test'")
		}

		return nil
	}
}

func TestCloneBuilderAcc_windows(t *testing.T) {
	t.Skip("Acceptance tests not configured yet.")
	t.Skip("test is too slow")
	config := windowsConfig()
	builderT.Test(t, builderT.TestCase{
		Builder:  &Builder{},
		Template: commonT.RenderConfig(config),
	})
}

func windowsConfig() map[string]interface{} {
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

		"vm_name":      commonT.NewVMName(),
		"template":     "windows",
		"host":         "esxi-1.vsphere65.test",
		"linked_clone": true, // speed up

		"communicator":   "winrm",
		"winrm_username": "jetbrains",
		"winrm_password": "jetbrains",
	}

	return config
}

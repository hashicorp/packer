package main

import (
	"testing"
	builderT "github.com/hashicorp/packer/helper/builder/testing"
	"github.com/hashicorp/packer/packer"
	"encoding/json"
	"github.com/jetbrains-infra/packer-builder-vsphere/driver"
	driverT "github.com/jetbrains-infra/packer-builder-vsphere/driver/testing"
)

func TestBuilderAcc_default(t *testing.T) {
	config := defaultConfig()
	builderT.Test(t, builderT.TestCase{
		Builder:  &Builder{},
		Template: renderConfig(config),
		Check:    checkDefault(t, config["vm_name"].(string), config["host"].(string), "datastore1"),
	})
}

func defaultConfig() map[string]interface{} {
	config := map[string]interface{}{
		"vcenter_server":      driverT.DefaultVCenterServer,
		"username":            driverT.DefaultVCenterUsername,
		"password":            driverT.DefaultVCenterPassword,
		"insecure_connection": true,

		"template": driverT.DefaultTemplate,
		"host":     driverT.DefaultHost,

		"ssh_username": "root",
		"ssh_password": "jetbrains",
	}
	config["vm_name"] = driverT.NewVMName()
	return config
}

func checkDefault(t *testing.T, name string, host string, datastore string) builderT.TestCheckFunc {
	return func(artifacts []packer.Artifact) error {
		d := driverT.NewTestDriver(t)
		vm := getVM(t, d, artifacts)
		return driverT.VMCheckDefault(t, d, vm, name, host, datastore)
	}
}

func TestBuilderAcc_artifact(t *testing.T) {
	config := defaultConfig()
	builderT.Test(t, builderT.TestCase{
		Builder:  &Builder{},
		Template: renderConfig(config),
		Check:    checkArtifact(t),
	})
}

func checkArtifact(t *testing.T) builderT.TestCheckFunc {
	return func(artifacts []packer.Artifact) error {
		if len(artifacts) > 1 {
			t.Fatal("more than 1 artifact")
		}

		artifactRaw := artifacts[0]
		_, ok := artifactRaw.(*Artifact)
		if !ok {
			t.Fatalf("unknown artifact: %#v", artifactRaw)
		}

		return nil
	}
}

func TestBuilderAcc_folder(t *testing.T) {
	builderT.Test(t, builderT.TestCase{
		Builder:  &Builder{},
		Template: folderConfig(),
		Check:    checkFolder(t, "folder1/folder2"),
	})
}

func folderConfig() string {
	config := defaultConfig()
	config["folder"] = "folder1/folder2"
	config["linked_clone"] = true // speed up
	return renderConfig(config)
}

func checkFolder(t *testing.T, folder string) builderT.TestCheckFunc {
	return func(artifacts []packer.Artifact) error {
		d := driverT.NewTestDriver(t)
		vm := getVM(t, d, artifacts)

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

func TestBuilderAcc_resourcePool(t *testing.T) {
	builderT.Test(t, builderT.TestCase{
		Builder:  &Builder{},
		Template: resourcePoolConfig(),
		Check:    checkResourcePool(t, "pool1/pool2"),
	})
}

func resourcePoolConfig() string {
	config := defaultConfig()
	config["resource_pool"] = "pool1/pool2"
	config["linked_clone"] = true // speed up
	return renderConfig(config)
}

func checkResourcePool(t *testing.T, pool string) builderT.TestCheckFunc {
	return func(artifacts []packer.Artifact) error {
		d := driverT.NewTestDriver(t)
		vm := getVM(t, d, artifacts)

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

// FIXME: why do we need this??? Why don't perform these checks in checkDefault?
func TestBuilderAcc_datastore(t *testing.T) {
	builderT.Test(t, builderT.TestCase{
		Builder:  &Builder{},
		Template: datastoreConfig(),
		Check:    checkDatastore(t, "datastore1"), // on esxi-1.vsphere55.test
	})
}

func datastoreConfig() string {
	config := defaultConfig()
	config["template"] = "alpine-host4" // on esxi-4.vsphere55.test
	return renderConfig(config)
}

func checkDatastore(t *testing.T, name string) builderT.TestCheckFunc {
	return func(artifacts []packer.Artifact) error {
		d := driverT.NewTestDriver(t)
		vm := getVM(t, d, artifacts)

		vmInfo, err := vm.Info("datastore")
		if err != nil {
			t.Fatalf("Cannot read VM properties: %v", err)
		}

		n := len(vmInfo.Datastore)
		if n != 1 {
			t.Fatalf("VM should have 1 datastore, got %v", n)
		}

		ds := d.NewDatastore(&vmInfo.Datastore[0])
		driverT.CheckDatastoreName(t, ds, name)

		return nil
	}
}

func TestBuilderAcc_multipleDatastores(t *testing.T) {
	t.Skip("test must fail") // FIXME

	builderT.Test(t, builderT.TestCase{
		Builder:  &Builder{},
		Template: multipleDatastoresConfig(),
	})
}

func multipleDatastoresConfig() string {
	config := defaultConfig()
	config["host"] = "esxi-4.vsphere55.test" // host with 2 datastores
	return renderConfig(config)
}

func TestBuilderAcc_linkedClone(t *testing.T) {
	builderT.Test(t, builderT.TestCase{
		Builder:  &Builder{},
		Template: linkedCloneConfig(),
		Check:    checkLinkedClone(t),
	})
}

func linkedCloneConfig() string {
	config := defaultConfig()
	config["linked_clone"] = true
	return renderConfig(config)
}

func checkLinkedClone(t *testing.T) builderT.TestCheckFunc {
	return func(artifacts []packer.Artifact) error {
		d := driverT.NewTestDriver(t)
		vm := getVM(t, d, artifacts)

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

func TestBuilderAcc_hardware(t *testing.T) {
	builderT.Test(t, builderT.TestCase{
		Builder:  &Builder{},
		Template: hardwareConfig(),
		Check:    checkHardware(t),
	})
}

func hardwareConfig() string {
	config := defaultConfig()
	config["CPUs"] = driverT.DefaultCPUs
	config["CPU_reservation"] = driverT.DefaultCPUReservation
	config["CPU_limit"] = driverT.DefaultCPULimit
	config["RAM"] = driverT.DefaultRAM
	config["RAM_reservation"] = driverT.DefaultRAMReservation
	config["linked_clone"] = true // speed up

	return renderConfig(config)
}

func checkHardware(t *testing.T) builderT.TestCheckFunc {
	return func(artifacts []packer.Artifact) error {
		d := driverT.NewTestDriver(t)
		vm := getVM(t, d, artifacts)
		return driverT.VMCheckHardware(t, d, vm)
	}
}

func TestBuilderAcc_RAMReservation(t *testing.T) {
	builderT.Test(t, builderT.TestCase{
		Builder:  &Builder{},
		Template: RAMReservationConfig(),
		Check:    checkRAMReservation(t),
	})
}

func RAMReservationConfig() string {
	config := defaultConfig()
	config["RAM_reserve_all"] = true
	config["linked_clone"] = true // speed up

	return renderConfig(config)
}

func checkRAMReservation(t *testing.T) builderT.TestCheckFunc {
	return func(artifacts []packer.Artifact) error {
		d := driverT.NewTestDriver(t)

		vm := getVM(t, d, artifacts)
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

func TestBuilderAcc_sshKey(t *testing.T) {
	builderT.Test(t, builderT.TestCase{
		Builder:  &Builder{},
		Template: sshKeyConfig(),
	})
}

func sshKeyConfig() string {
	config := defaultConfig()
	config["ssh_password"] = ""
	config["ssh_private_key_file"] = "test-key.pem"
	config["linked_clone"] = true // speed up
	return renderConfig(config)
}

func TestBuilderAcc_snapshot(t *testing.T) {
	builderT.Test(t, builderT.TestCase{
		Builder:  &Builder{},
		Template: snapshotConfig(),
		Check:    checkSnapshot(t),
	})
}

func snapshotConfig() string {
	config := defaultConfig()
	config["create_snapshot"] = true
	return renderConfig(config)
}

func checkSnapshot(t *testing.T) builderT.TestCheckFunc {
	return func(artifacts []packer.Artifact) error {
		d := driverT.NewTestDriver(t)

		vm := getVM(t, d, artifacts)
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

func TestBuilderAcc_template(t *testing.T) {
	builderT.Test(t, builderT.TestCase{
		Builder:  &Builder{},
		Template: templateConfig(),
		Check:    checkTemplate(t),
	})
}

func templateConfig() string {
	config := defaultConfig()
	config["convert_to_template"] = true
	config["linked_clone"] = true // speed up
	return renderConfig(config)
}

func checkTemplate(t *testing.T) builderT.TestCheckFunc {
	return func(artifacts []packer.Artifact) error {
		d := driverT.NewTestDriver(t)

		vm := getVM(t, d, artifacts)
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

func renderConfig(config map[string]interface{}) string {
	t := map[string][]map[string]interface{}{
		"builders": {
			map[string]interface{}{
				"type": "test",
			},
		},
	}
	for k, v := range config {
		t["builders"][0][k] = v
	}

	j, _ := json.Marshal(t)
	return string(j)
}

func getVM(t *testing.T, d *driver.Driver, artifacts []packer.Artifact) *driver.VirtualMachine {
	artifactRaw := artifacts[0]
	artifact, _ := artifactRaw.(*Artifact)

	vm, err := d.FindVM(artifact.Name)
	if err != nil {
		t.Fatalf("Cannot find VM: %v", err)
	}

	return vm
}

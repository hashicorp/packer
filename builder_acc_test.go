package main

import (
	"encoding/json"
	builderT "github.com/hashicorp/packer/helper/builder/testing"
	"github.com/hashicorp/packer/packer"
	"github.com/jetbrains-infra/packer-builder-vsphere/driver"
	driverT "github.com/jetbrains-infra/packer-builder-vsphere/driver/testing"
	"testing"
)

func TestBuilderAcc_default(t *testing.T) {
	config := defaultConfig()
	builderT.Test(t, builderT.TestCase{
		Builder:  &Builder{},
		Template: renderConfig(config),
		Check: func(artifacts []packer.Artifact) error {
			d := driverT.NewTestDriver(t)
			driverT.VMCheckDefault(t, d, getVM(t, d, artifacts), config["vm_name"].(string),
				config["host"].(string), driverT.TestDatastore)
			return nil
		},
	})
}

func defaultConfig() map[string]interface{} {
	config := map[string]interface{}{
		"vcenter_server":      driverT.TestVCenterServer,
		"username":            driverT.TestVCenterUsername,
		"password":            driverT.TestVCenterPassword,
		"insecure_connection": true,

		"template": driverT.TestTemplate,
		"host":     driverT.TestHost,

		"ssh_username": "root",
		"ssh_password": "jetbrains",
	}
	config["vm_name"] = driverT.NewVMName()
	return config
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
		Check:    checkFolder(t, driverT.TestFolder),
	})
}

func folderConfig() string {
	config := defaultConfig()
	config["folder"] = driverT.TestFolder
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
		driverT.CheckFolderPath(t, f, folder)

		return nil
	}
}

func TestBuilderAcc_resourcePool(t *testing.T) {
	builderT.Test(t, builderT.TestCase{
		Builder:  &Builder{},
		Template: resourcePoolConfig(),
		Check:    checkResourcePool(t, driverT.TestResourcePool),
	})
}

func resourcePoolConfig() string {
	config := defaultConfig()
	config["resource_pool"] = driverT.TestResourcePool
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

		driverT.CheckResourcePoolPath(t, d.NewResourcePool(vmInfo.ResourcePool), pool)
		return nil
	}
}

func TestBuilderAcc_datastore(t *testing.T) {
	builderT.Test(t, builderT.TestCase{
		Builder:  &Builder{},
		Template: datastoreConfig(),
		Check:    func(artifacts []packer.Artifact) error {
			d := driverT.NewTestDriver(t)
			driverT.VMCheckDatastore(t, d, getVM(t, d, artifacts), driverT.TestDatastore)
			return nil
		},
	})
}

func datastoreConfig() string {
	config := defaultConfig()
	config["template"] = "alpine-host4" // on esxi-4.vsphere55.test
	return renderConfig(config)
}

func TestBuilderAcc_multipleDatastores(t *testing.T) {
	t.Skip("test must fail")

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
		Check:    func(artifacts []packer.Artifact) error {
			d := driverT.NewTestDriver(t)
			driverT.VMCheckHardware(t, d, getVM(t, d, artifacts))
			return nil
		},
	})
}

func hardwareConfig() string {
	config := defaultConfig()
	config["CPUs"] = driverT.TestCPUs
	config["CPU_reservation"] = driverT.TestCPUReservation
	config["CPU_limit"] = driverT.TestCPULimit
	config["RAM"] = driverT.TestRAM
	config["RAM_reservation"] = driverT.TestRAMReservation
	config["linked_clone"] = true // speed up

	return renderConfig(config)
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
		Check:    func(artifacts []packer.Artifact) error {
			d := driverT.NewTestDriver(t)
			driverT.VMCheckSnapshor(t, d, getVM(t, d, artifacts))
			return nil
		},
	})
}

func snapshotConfig() string {
	config := defaultConfig()
	config["create_snapshot"] = true
	return renderConfig(config)
}

func TestBuilderAcc_template(t *testing.T) {
	builderT.Test(t, builderT.TestCase{
		Builder:  &Builder{},
		Template: templateConfig(),
		Check: func(artifacts []packer.Artifact) error {
			d := driverT.NewTestDriver(t)
			driverT.VMCheckTemplate(t, d, getVM(t, d, artifacts))
			return nil
		},
	})
}

func templateConfig() string {
	config := defaultConfig()
	config["convert_to_template"] = true
	config["linked_clone"] = true // speed up
	return renderConfig(config)
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

package main

import (
	"testing"
	builderT "github.com/hashicorp/packer/helper/builder/testing"
	"fmt"
	"github.com/hashicorp/packer/packer"
	"encoding/json"
	"math/rand"
	"github.com/vmware/govmomi/object"
	"github.com/jetbrains-infra/packer-builder-vsphere/driver"
)

func TestBuilderAcc_default(t *testing.T) {
	config := defaultConfig()
	builderT.Test(t, builderT.TestCase{
		Builder:  &Builder{},
		Template: renderConfig(config),
		Check:    checkDefault(t, config["vm_name"].(string), config["host"].(string)),
	})
}

func defaultConfig() map[string]interface{} {
	config := map[string]interface{}{
		"vcenter_server":      "vcenter.vsphere55.test",
		"username":            "root",
		"password":            "jetbrains",
		"insecure_connection": true,

		"template": "basic",
		"host":     "esxi-1.vsphere55.test",

		"ssh_username": "jetbrains",
		"ssh_password": "jetbrains",
	}
	config["vm_name"] = fmt.Sprintf("test-%v", rand.Intn(1000))
	return config
}

func checkDefault(t *testing.T, name string, host string) builderT.TestCheckFunc {
	return func(artifacts []packer.Artifact) error {
		d := testConn(t)
		vm := getVM(t, d, artifacts)

		vmInfo, err := d.VMInfo(vm, "name", "parent", "runtime.host", "resourcePool", "layoutEx.disk")
		if err != nil {
			t.Fatalf("Cannot read VM properties: %v", err)
		}

		if vmInfo.Name != name {
			t.Errorf("Invalid VM name: expected '%v', got '%v'", name, vmInfo.Name)
		}

		f := d.NewFolder(vmInfo.Parent)
		folderPath, err := d.GetFolderPath(f)
		if err != nil {
			t.Fatalf("Cannot read folder name: %v", err)
		}
		if folderPath != "" {
			t.Errorf("Invalid folder: expected '/', got '%v'", folderPath)
		}

		h := d.NewHost(vmInfo.Runtime.Host)
		hostInfo, err := d.HostInfo(h, "name")
		if err != nil {
			t.Fatal("Cannot read host properties: ", err)
		}

		if hostInfo.Name != host {
			t.Errorf("Invalid host name: expected '%v', got '%v'", host, hostInfo.Name)
		}

		p := d.NewResourcePool(vmInfo.ResourcePool)
		poolPath, err := d.GetResourcePoolPath(p)
		if err != nil {
			t.Fatalf("Cannot read resource pool name: %v", err)
		}
		if poolPath != "" {
			t.Error("Invalid resource pool: expected '/', got '%v'", poolPath)
		}

		if len(vmInfo.LayoutEx.Disk[0].Chain) != 1 {
			t.Error("Not a full clone")
		}

		return nil
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
		d := testConn(t)
		vm := getVM(t, d, artifacts)

		vmInfo, err := d.VMInfo(vm, "parent")
		if err != nil {
			t.Fatalf("Cannot read VM properties: %v", err)
		}

		f := d.NewFolder(vmInfo.Parent)
		path, err := d.GetFolderPath(f)
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
		d := testConn(t)
		vm := getVM(t, d, artifacts)

		vmInfo, err := d.VMInfo(vm, "resourcePool")
		if err != nil {
			t.Fatalf("Cannot read VM properties: %v", err)
		}

		p := d.NewResourcePool(vmInfo.ResourcePool)
		path, err := d.GetResourcePoolPath(p)
		if err != nil {
			t.Fatalf("Cannot read resource pool name: %v", err)
		}
		if path != pool {
			t.Errorf("Wrong folder. expected: %v, got: %v", pool, path)
		}

		return nil
	}
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
		d := testConn(t)
		vm := getVM(t, d, artifacts)

		vmInfo, err := d.VMInfo(vm, "layoutEx.disk")
		if err != nil {
			t.Fatalf("Cannot read VM properties: %v", err)
		}

		if len(vmInfo.LayoutEx.Disk[0].Chain) != 3 {
			t.Error("Not a linked clone")
		}

		return nil
	}
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
		d := testConn(t)

		vm := getVM(t, d, artifacts)
		vmInfo, err := d.VMInfo(vm, "layoutEx.disk")
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
		d := testConn(t)

		vm := getVM(t, d, artifacts)
		vmInfo, err := d.VMInfo(vm, "config.template")
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

func testConn(t *testing.T) *driver.Driver {
	d, err := driver.NewDriver(&driver.ConnectConfig{
		VCenterServer:      "vcenter.vsphere55.test",
		Username:           "root",
		Password:           "jetbrains",
		InsecureConnection: true,
	})
	if err != nil {
		t.Fatal("Cannot connect: ", err)
	}
	return d
}

func getVM(t *testing.T, d *driver.Driver, artifacts []packer.Artifact) *object.VirtualMachine {
	artifactRaw := artifacts[0]
	artifact, _ := artifactRaw.(*Artifact)

	vm, err := d.FindVM(artifact.Name)
	if err != nil {
		t.Fatalf("Cannot find VM: %v", err)
	}

	return vm
}

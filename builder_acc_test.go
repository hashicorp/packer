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
		if len(artifacts) > 1 {
			t.Fatal("more than 1 artifact")
		}

		artifactRaw := artifacts[0]
		artifact, ok := artifactRaw.(*Artifact)
		if !ok {
			t.Fatalf("unknown artifact: %#v", artifactRaw)
		}

		conn := testConn(t)

		vm, err := conn.FindVM(artifact.Name)
		if err != nil {
			t.Fatal("Cannot find VM: ", err)
		}

		vmInfo, err := conn.VMInfo(vm, "name", "runtime.host", "resourcePool", "layoutEx.disk")
		if err != nil {
			t.Fatalf("Cannot read VM properties: %v", err)
		}

		if vmInfo.Name != name {
			t.Errorf("Invalid VM name: expected '%v', got '%v'", name, vmInfo.Name)
		}

		h := conn.NewHost(vmInfo.Runtime.Host)
		hostInfo, err := conn.HostInfo(h, "name")
		if err != nil {
			t.Fatal("Cannot read VM properties: ", err)
		}

		if hostInfo.Name != host {
			t.Errorf("Invalid host name: expected '%v', got '%v'", host, hostInfo.Name)
		}

		p := conn.NewResourcePool(vmInfo.ResourcePool)
		poolInfo, err := conn.ResourcePoolInfo(p, "owner", "parent")
		if err != nil {
			t.Fatalf("Cannot read resource pool properties: %v", err)
		}

		if poolInfo.Owner != *poolInfo.Parent {
			t.Error("Not a root resource pool")
		}

		if len(vmInfo.LayoutEx.Disk[0].Chain) != 1 {
			t.Error("Not a full clone")
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
		artifactRaw := artifacts[0]
		artifact, _ := artifactRaw.(*Artifact)

		conn := testConn(t)

		vm, err := conn.FindVM(artifact.Name)
		if err != nil {
			t.Fatalf("Cannot find VM: %v", err)
		}

		vmInfo, err := conn.VMInfo(vm, "layoutEx.disk")
		if err != nil {
			t.Fatalf("Cannot read VM properties: %v", err)
		}

		if len(vmInfo.LayoutEx.Disk[0].Chain) != 3 {
			t.Error("Not a linked clone")
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
	config["linked_clone"] = true
	config["convert_to_template"] = true
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

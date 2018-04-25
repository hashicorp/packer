package testing

import (
	"fmt"
	"math/rand"
	"time"
	"encoding/json"
	"github.com/hashicorp/packer/packer"
	"github.com/jetbrains-infra/packer-builder-vsphere/driver"
	"testing"
	"github.com/jetbrains-infra/packer-builder-vsphere/common"
	"context"
)

func NewVMName() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("test-%v", rand.Intn(1000))
}

func RenderConfig(config map[string]interface{}) string {
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


func TestConn(t *testing.T) *driver.Driver {
	d, err := driver.NewDriver(context.TODO(), &driver.ConnectConfig{
		VCenterServer:      "vcenter.vsphere65.test",
		Username:           "root",
		Password:           "jetbrains",
		InsecureConnection: true,
	})
	if err != nil {
		t.Fatal("Cannot connect: ", err)
	}
	return d
}

func GetVM(t *testing.T, d *driver.Driver, artifacts []packer.Artifact) *driver.VirtualMachine {
	artifactRaw := artifacts[0]
	artifact, _ := artifactRaw.(*common.Artifact)

	vm, err := d.FindVM(artifact.Name)
	if err != nil {
		t.Fatalf("Cannot find VM: %v", err)
	}

	return vm
}


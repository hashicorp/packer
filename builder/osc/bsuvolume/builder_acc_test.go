//TODO: explain how to delete the image.
package bsuvolume

import (
	"crypto/tls"
	"net/http"
	"testing"

	"github.com/hashicorp/packer/builder/osc/common"
	builderT "github.com/hashicorp/packer/helper/builder/testing"
	"github.com/outscale/osc-go/oapi"
)

func TestBuilderAcc_basic(t *testing.T) {
	builderT.Test(t, builderT.TestCase{
		PreCheck:             func() { testAccPreCheck(t) },
		Builder:              &Builder{},
		Template:             testBuilderAccBasic,
		SkipArtifactTeardown: true,
	})
}

func testAccPreCheck(t *testing.T) {
}

func testOAPIConn() (*oapi.Client, error) {
	access := &common.AccessConfig{RawRegion: "us-east-1"}
	clientConfig, err := access.Config()
	if err != nil {
		return nil, err
	}

	skipClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	return oapi.NewClient(clientConfig, skipClient), nil
}

const testBuilderAccBasic = `
{
    "builders": [
        {
            "type": "test",
            "region": "eu-west-2",
            "vm_type": "t2.micro",
            "source_omi": "ami-65efcc11",
            "ssh_username": "outscale",
            "bsu_volumes": [
                {
                    "volume_type": "gp2",
                    "device_name": "/dev/xvdf",
                    "delete_on_vm_deletion": false,
                    "tags": {
                        "zpool": "data",
                        "Name": "Data1"
                    },
                    "volume_size": 10
                },
                {
                    "volume_type": "gp2",
                    "device_name": "/dev/xvdg",
                    "tags": {
                        "zpool": "data",
                        "Name": "Data2"
                    },
                    "delete_on_vm_deletion": false,
                    "volume_size": 10
                },
                {
                    "volume_size": 10,
                    "tags": {
                        "Name": "Data3",
                        "zpool": "data"
                    },
                    "delete_on_vm_deletion": false,
                    "device_name": "/dev/xvdh",
                    "volume_type": "gp2"
                }
            ]
        }
    ]
}
`

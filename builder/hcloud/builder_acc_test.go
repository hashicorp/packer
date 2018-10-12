package hcloud

import (
	"os"
	"testing"

	builderT "github.com/hashicorp/packer/helper/builder/testing"
)

func TestBuilderAcc_basic(t *testing.T) {
	builderT.Test(t, builderT.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Builder:  &Builder{},
		Template: testBuilderAccBasic,
	})
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("HCLOUD_TOKEN"); v == "" {
		t.Fatal("HCLOUD_TOKEN must be set for acceptance tests")
	}
}

const testBuilderAccBasic = `
{
	"builders": [{
		"type": "test",
		"location": "nbg1",
		"server_type": "cx11",
		"image": "ubuntu-18.04",
		"user_data": "",
		"user_data_file": ""
	}]
}
`

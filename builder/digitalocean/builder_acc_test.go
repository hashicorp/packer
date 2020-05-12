package digitalocean

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
	if v := os.Getenv("DIGITALOCEAN_API_TOKEN"); v == "" {
		t.Fatal("DIGITALOCEAN_API_TOKEN must be set for acceptance tests")
	}
}

const testBuilderAccBasic = `
{
	"builders": [{
		"type": "test",
		"region": "nyc2",
		"size": "s-1vcpu-1gb",
		"image": "ubuntu-20-04-x64",
		"ssh_username": "root",
		"user_data": "",
		"user_data_file": ""
	}]
}
`

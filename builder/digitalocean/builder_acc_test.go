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
		"region": "nyc1",
		"size": "512mb",
		"image": "ubuntu-16-04-x64",
		"ssh_username": "root",
		"volumes": [
			{ "size": 5 },
			{ "size": 10 }
		]
	}]
}
`

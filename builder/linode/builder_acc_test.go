package linode

import (
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"os"
	"testing"

	builderT "github.com/hashicorp/packer-plugin-sdk/acctest"
)

func TestBuilderAcc_multiple(t *testing.T) {
	builderT.Test(t, builderT.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Builder:  &Builder{},
		Template: testBuilderAccBasic,
		SkipArtifactTeardown: true,
	})
	builderT.Test(t, builderT.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Builder:  &Builder{},
		Template: testBuilderAccLimited,
		Teardown: func () { cleanTestImages(t) },
	})
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("LINODE_TOKEN"); v == "" {
		t.Fatal("LINODE_TOKEN must be set for acceptance tests")
	}
}

func cleanTestImages(t *testing.T) {
}

const testBuilderAccBasic = `
{
	"builders": [{
		"type": "test",
		"region": "us-east",
		"instance_type": "g6-nanode-1",
		"image": "linode/alpine3.9",
		"ssh_username": "root"
	}]
}
`
const testBuilderAccLimited = `
{
	"builders": [
		{
			"type": "test",
			"region": "us-east",
			"instance_type": "g6-nanode-1",
			"image": "linode/alpine3.9",
			"ssh_username": "root",
			"account_image_limit": 1
		}
	]
}
`

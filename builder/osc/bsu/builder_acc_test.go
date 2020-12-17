//TODO: explain how to delete the image.
package bsu

import (
	"testing"

	builderT "github.com/hashicorp/packer-plugin-sdk/acctest"
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

const testBuilderAccBasic = `
{
	"builders": [{
		"type": "test",
		"region": "eu-west-2",
		"vm_type": "t2.micro",
		"source_omi": "ami-abe953fa",
		"ssh_username": "outscale",
		"omi_name": "packer-test",
		"associate_public_ip_address": true,
		"force_deregister": true
	}]
}
`

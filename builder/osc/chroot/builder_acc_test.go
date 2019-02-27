package chroot

import (
	"testing"

	builderT "github.com/hashicorp/packer/helper/builder/testing"
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
		"source_omi": "ami-99466096",
		"omi_name": "packer-test {{timestamp}}",
		"omi_virtualization_type": "hvm",
		"device_path": "/dev/xvdf",
		"mount_partition": "0"
	}]
}
`

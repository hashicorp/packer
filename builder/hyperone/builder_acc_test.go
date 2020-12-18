package hyperone

import (
	"os"
	"testing"

	builderT "github.com/hashicorp/packer-plugin-sdk/acctest"
)

func TestBuilderAcc_basic(t *testing.T) {
	builderT.Test(t, builderT.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Builder:  &Builder{},
		Template: testBuilderAccBasic,
	})
}

func TestBuilderAcc_chroot(t *testing.T) {
	builderT.Test(t, builderT.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Builder:  &Builder{},
		Template: testBuilderAccChroot,
	})
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("HYPERONE_TOKEN"); v == "" {
		t.Fatal("HYPERONE_TOKEN must be set for acceptance tests")
	}
}

const testBuilderAccBasic = `
{
	"builders": [{
		"type": "test",
		"vm_type": "a1.nano",
		"source_image": "ubuntu",
		"disk_size": 10,
		"image_tags": {
			"key":"value"
		},
		"vm_tags": {
			"key_vm":"value_vm"
		}
	}]
}
`

const testBuilderAccChroot = `
{
	"builders": [{
		"type": "test",
		"source_image": "ubuntu",
		"disk_size": 10,
		"vm_type": "a1.nano",
		"chroot_disk": true,
		"chroot_command_wrapper": "sudo {{.Command}}",
		"pre_mount_commands": [
			"parted {{.Device}} mklabel msdos mkpart primary 1M 100% set 1 boot on print",
			"mkfs.ext4 {{.Device}}1"
		],
		"post_mount_commands": [
			"apt-get update",
			"apt-get install debootstrap",
			"debootstrap --arch amd64 bionic {{.MountPath}}"
		]
	}]
}
`

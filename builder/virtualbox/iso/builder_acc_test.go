package iso

import (
	"testing"

	builderT "github.com/mitchellh/packer/helper/builder/testing"
)

func TestBuilderAcc_basic(t *testing.T) {
	builderT.Test(t, builderT.TestCase{
		Builder:  &Builder{},
		Template: testBuilderAccBasic,
	})
}

const testBuilderAccBasic = `
{
	"builders": [{
		"type": "test",
		"guest_os_type": "Ubuntu_64",
		"iso_url": "http://releases.ubuntu.com/12.04/ubuntu-12.04.5-server-amd64.iso",
		"iso_checksum": "769474248a3897f4865817446f9a4a53",
		"iso_checksum_type": "md5",
		"ssh_username": "packer",
		"ssh_password": "packer",
		"ssh_wait_timeout": "30s",
		"shutdown_command": "echo 'packer' | sudo -S shutdown -P now"
	}]
}
`

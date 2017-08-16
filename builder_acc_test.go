package main

import (
	"testing"
	builderT "github.com/hashicorp/packer/helper/builder/testing"
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

		"vcenter_server": "vcenter.vsphere55.test",
		"username": "root",
		"password": "jetbrains",
		"insecure_connection": true,

		"template": "basic",
		"vm_name": "test1",
		"host": "esxi-1.vsphere55.test",

		"ssh_username": "jetbrains",
		"ssh_password": "jetbrains"
	}]
}
`

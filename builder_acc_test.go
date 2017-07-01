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

		"vcenter_host": "vcenter.vsphere5.test",
		"username": "root",
		"password": "jetbrains",

		"template": "basic",
		"vm_name": "test1",
		"host": "esxi-1.vsphere5.test",

		"ssh_username": "jetbrains",
		"ssh_password": "jetbrains"
	}]
}
`

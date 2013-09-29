package main

import (
	"github.com/mitchellh/packer/packer/plugin"
	"github.com/mitchellh/packer/provisioner/ansible-local"
)

func main() {
	plugin.ServeProvisioner(new(ansiblelocal.Provisioner))
}

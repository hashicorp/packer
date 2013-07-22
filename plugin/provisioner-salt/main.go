package main

import (
	"github.com/mitchellh/packer/packer/plugin"
	"github.com/mitchellh/packer/provisioner/salt"
)

func main() {
	plugin.ServeProvisioner(new(salt.Provisioner))
}

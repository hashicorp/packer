package main

import (
	"github.com/mitchellh/packer/packer/plugin"
	"github.com/mitchellh/packer/provisioner/salt-masterless"
)

func main() {
	plugin.ServeProvisioner(new(saltmasterless.Provisioner))
}

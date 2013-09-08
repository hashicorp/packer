package main

import (
	"github.com/mitchellh/packer/packer/plugin"
	"github.com/mitchellh/packer/provisioner/puppet-masterless"
)

func main() {
	plugin.ServeProvisioner(new(puppetmasterless.Provisioner))
}

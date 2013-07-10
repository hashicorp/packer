package main

import (
	"github.com/jvandyke/packer/provisioner/chef-solo"
	"github.com/mitchellh/packer/packer/plugin"
)

func main() {
	plugin.ServeProvisioner(new(chefSolo.Provisioner))
}

package main

import (
	"github.com/mitchellh/packer/packer/plugin"
	"../../provisioner/chef-solo"
)

func main() {
	plugin.ServeProvisioner(new(chefSolo.Provisioner))
}

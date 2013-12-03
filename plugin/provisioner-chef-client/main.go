package main

import (
	"github.com/mitchellh/packer/packer/plugin"
	"github.com/mitchellh/packer/provisioner/chef-client"
)

func main() {
	plugin.ServeProvisioner(new(chefclient.Provisioner))
}

package main

import (
	"github.com/mitchellh/packer/packer/plugin"
	"github.com/mitchellh/packer/provisioner/chef-solo"
)

func main() {
	plugin.ServeProvisioner(new(chefsolo.Provisioner))
}

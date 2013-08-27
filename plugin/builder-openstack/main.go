package main

import (
	"github.com/mitchellh/packer/builder/openstack"
	"github.com/mitchellh/packer/packer/plugin"
)

func main() {
	plugin.ServeBuilder(new(openstack.Builder))
}

package main

import (
	"github.com/mitchellh/packer/builder/vsphere"
	"github.com/mitchellh/packer/packer/plugin"
)

func main() {
	plugin.ServeBuilder(new(vsphere.Builder))
}

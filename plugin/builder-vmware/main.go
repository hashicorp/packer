package main

import (
	"github.com/mitchellh/packer/builder/vmware"
	"github.com/mitchellh/packer/packer/plugin"
)

func main() {
	plugin.ServeBuilder(new(vmware.Builder))
}

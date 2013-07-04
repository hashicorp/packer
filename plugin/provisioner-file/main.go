package main

import (
	"github.com/mitchellh/packer/packer/plugin"
	"github.com/mitchellh/packer/provisioner/file"
)

func main() {
	plugin.ServeProvisioner(new(file.Provisioner))
}

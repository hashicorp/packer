package main

import (
	"github.com/mitchellh/packer/packer/plugin"
	"github.com/mitchellh/packer/provisioner/shell"
)

func main() {
	plugin.ServeProvisioner(new(shell.Provisioner))
}

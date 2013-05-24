package main

import (
	"github.com/mitchellh/packer/provisioner/shell"
	"github.com/mitchellh/packer/packer/plugin"
)

func main() {
	plugin.ServeProvisioner(new(shell.Provisioner))
}

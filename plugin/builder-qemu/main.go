package main

import (
	"github.com/mitchellh/packer/builder/qemu"
	"github.com/mitchellh/packer/packer/plugin"
)

func main() {
	plugin.ServeBuilder(new(qemu.Builder))
}

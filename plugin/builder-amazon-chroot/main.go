package main

import (
	"github.com/mitchellh/packer/builder/amazon/chroot"
	"github.com/mitchellh/packer/packer/plugin"
)

func main() {
	plugin.ServeBuilder(new(chroot.Builder))
}

package main

import (
	"github.com/mitchellh/packer/builder/virtualbox"
	"github.com/mitchellh/packer/packer/plugin"
)

func main() {
	plugin.ServeBuilder(new(virtualbox.Builder))
}

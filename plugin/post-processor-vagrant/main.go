package main

import (
	"github.com/mitchellh/packer/packer/plugin"
	"github.com/mitchellh/packer/post-processor/vagrant"
)

func main() {
	plugin.ServePostProcessor(new(vagrant.PostProcessor))
}

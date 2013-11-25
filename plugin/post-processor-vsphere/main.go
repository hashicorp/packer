package main

import (
	"github.com/mitchellh/packer/packer/plugin"
	"github.com/mitchellh/packer/post-processor/vsphere"
)

func main() {
	plugin.ServePostProcessor(new(vsphere.PostProcessor))
}

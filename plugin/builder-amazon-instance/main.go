package main

import (
	"github.com/mitchellh/packer/builder/amazon/instance"
	"github.com/mitchellh/packer/packer/plugin"
)

func main() {
	plugin.ServeBuilder(new(instance.Builder))
}

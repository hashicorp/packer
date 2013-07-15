package main

import (
	"github.com/mitchellh/packer/builder/amazon/ebs"
	"github.com/mitchellh/packer/packer/plugin"
)

func main() {
	plugin.ServeBuilder(new(ebs.Builder))
}

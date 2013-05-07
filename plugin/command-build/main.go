package main

import (
	"github.com/mitchellh/packer/command/build"
	"github.com/mitchellh/packer/packer/plugin"
)

func main() {
	plugin.ServeCommand(new(build.Command))
}

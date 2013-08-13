package main

import (
	"github.com/mitchellh/packer/command/inspect"
	"github.com/mitchellh/packer/packer/plugin"
)

func main() {
	plugin.ServeCommand(new(inspect.Command))
}

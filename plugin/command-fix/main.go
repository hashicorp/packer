package main

import (
	"github.com/mitchellh/packer/command/fix"
	"github.com/mitchellh/packer/packer/plugin"
)

func main() {
	plugin.ServeCommand(new(fix.Command))
}

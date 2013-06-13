package main

import (
	"github.com/mitchellh/packer/command/validate"
	"github.com/mitchellh/packer/packer/plugin"
)

func main() {
	plugin.ServeCommand(new(validate.Command))
}

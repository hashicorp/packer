package main

import (
	"github.com/mitchellh/packer/command/fix"
	"github.com/mitchellh/packer/packer/plugin"
)

func main() {
	server, err := plugin.Server()
	if err != nil {
		panic(err)
	}
	server.RegisterCommand(new(fix.Command))
	server.Serve()
}

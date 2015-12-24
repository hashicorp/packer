package main

import (
	"github.com/mitchellh/packer/builder/lxc"
	"github.com/mitchellh/packer/packer/plugin"
)

func main() {
	server, err := plugin.Server()
	if err != nil {
		panic(err)
	}
	server.RegisterBuilder(new(lxc.Builder))
	server.Serve()
}

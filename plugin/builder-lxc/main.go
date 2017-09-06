package main

import (
	"github.com/hashicorp/packer/builder/lxc"
	"github.com/hashicorp/packer/packer/plugin"
)

func main() {
	server, err := plugin.Server()
	if err != nil {
		panic(err)
	}
	server.RegisterBuilder(new(lxc.Builder))
	server.Serve()
}

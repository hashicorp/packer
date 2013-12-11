package main

import (
	"github.com/mitchellh/packer/builder/vmware"
	"github.com/mitchellh/packer/packer/plugin"
)

func main() {
	server, err := plugin.Server()
	if err != nil {
		panic(err)
	}
	server.RegisterBuilder(new(vmware.Builder))
	server.Serve()
}

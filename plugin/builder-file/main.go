package main

import (
	"github.com/mitchellh/packer/builder/file"
	"github.com/mitchellh/packer/packer/plugin"
)

func main() {
	server, err := plugin.Server()
	if err != nil {
		panic(err)
	}
	server.RegisterBuilder(new(file.Builder))
	server.Serve()
}

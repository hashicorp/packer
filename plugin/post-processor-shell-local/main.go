package main

import (
	"github.com/mitchellh/packer/post-processor/shell-local"
	"github.com/mitchellh/packer/packer/plugin"
)

func main() {
	server, err := plugin.Server()
	if err != nil {
		panic(err)
	}
	server.RegisterPostProcessor(new(shell_local.PostProcessor))
	server.Serve()
}

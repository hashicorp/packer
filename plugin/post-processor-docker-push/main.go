package main

import (
	"github.com/mitchellh/packer/packer/plugin"
	"github.com/mitchellh/packer/post-processor/docker-push"
)

func main() {
	server, err := plugin.Server()
	if err != nil {
		panic(err)
	}
	server.RegisterPostProcessor(new(dockerpush.PostProcessor))
	server.Serve()
}

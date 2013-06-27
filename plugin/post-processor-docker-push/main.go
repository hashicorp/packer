package main

import (
	"github.com/mitchellh/packer/packer/plugin"
	"github.com/mitchellh/packer/post-processor/docker-push"
)

func main() {
	plugin.ServePostProcessor(new(docker.PushPostProcessor))
}

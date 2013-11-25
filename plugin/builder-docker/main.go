package main

import (
	"github.com/mitchellh/packer/builder/docker"
	"github.com/mitchellh/packer/packer/plugin"
)

func main() {
	plugin.ServeBuilder(new(docker.Builder))
}

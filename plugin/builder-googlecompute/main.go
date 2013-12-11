package main

import (
	"github.com/kelseyhightower/packer-builder-googlecompute/builder/googlecompute"
	"github.com/mitchellh/packer/packer/plugin"
)

func main() {
	plugin.ServeBuilder(new(googlecompute.Builder))
}

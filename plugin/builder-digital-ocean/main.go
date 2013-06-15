package main

import (
	"github.com/mitchellh/packer/packer/plugin"
	"github.com/pearkes/packer/builder/digitalocean"
)

func main() {
	plugin.ServeBuilder(new(digitalocean.Builder))
}

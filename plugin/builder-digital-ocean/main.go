package main

import (
	"github.com/mitchellh/packer/packer/plugin"
	"github.com/mitchellh/packer/builder/digitalocean"
)

func main() {
	plugin.ServeBuilder(new(digitalocean.Builder))
}

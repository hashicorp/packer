package main

import (
	"github.com/mitchellh/packer/builder/digitalocean"
	"github.com/mitchellh/packer/packer/plugin"
)

func main() {
	plugin.ServeBuilder(new(digitalocean.Builder))
}

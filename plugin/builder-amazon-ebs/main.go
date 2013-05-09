package main

import (
	"github.com/mitchellh/packer/builder/amazonebs"
	"github.com/mitchellh/packer/packer/plugin"
)

func main() {
	plugin.ServeBuilder(new(amazonebs.Builder))
}

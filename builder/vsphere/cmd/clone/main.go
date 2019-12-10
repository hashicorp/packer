package main

import "github.com/hashicorp/packer/packer/plugin"
import "github.com/jetbrains-infra/packer-builder-vsphere/clone"

func main() {
	server, err := plugin.Server()
	if err != nil {
		panic(err)
	}
	err = server.RegisterBuilder(new(clone.Builder))
	if err != nil {
		panic(err)
	}
	server.Serve()
}

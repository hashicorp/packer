package main

import "github.com/hashicorp/packer/packer/plugin"
import "github.com/jetbrains-infra/packer-builder-vsphere/iso"

func main() {
	server, err := plugin.Server()
	if err != nil {
		panic(err)
	}
	server.RegisterBuilder(new(iso.Builder))
	server.Serve()
}

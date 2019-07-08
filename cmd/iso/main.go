package main

import "github.com/hashicorp/packer/packer/plugin"
import "github.com/jetbrains-infra/packer-builder-vsphere/iso"

func main() {
	server, err := plugin.Server()
	if err != nil {
		panic(err)
	}
	err = server.RegisterBuilder(new(iso.Builder))
	if err != nil {
		panic(err)
	}
	server.Serve()
}

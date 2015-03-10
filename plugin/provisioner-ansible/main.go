package main

import (
	"github.com/bhcleek/packer-provisioner-ansible/provisioner/ansible"
	"github.com/mitchellh/packer/packer/plugin"
)

func main() {
	server, err := plugin.Server()
	if err != nil {
		panic(err)
	}
	server.RegisterProvisioner(new(ansible.Provisioner))
	server.Serve()
}

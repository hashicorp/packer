package main

import (
  "github.com/mitchellh/packer/packer/plugin"
  "github.com/mitchellh/packer/provisioner/shell-local"
)

func main() {
  server, err := plugin.Server()
  if err != nil {
    panic(err)
  }
  server.RegisterProvisioner(new(shelllocal.Provisioner))
  server.Serve()
}

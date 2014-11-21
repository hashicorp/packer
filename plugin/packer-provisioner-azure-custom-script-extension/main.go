// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package main

import (
	"github.com/mitchellh/packer/packer/plugin"
	"github.com/mitchellh/packer/provisioner/azureVmCustomScriptExtension"
)

func main() {
	server, err := plugin.Server()
	if err != nil {
		panic(err)
	}
	server.RegisterProvisioner(new(azureVmCustomScriptExtension.Provisioner))
	server.Serve()
}

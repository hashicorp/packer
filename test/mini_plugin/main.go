// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"
	"os"

	"github.com/hashicorp/packer-plugin-sdk/plugin"
	"github.com/hashicorp/packer-plugin-tester/builder/dynamic"
	dynamicDS "github.com/hashicorp/packer-plugin-tester/datasource/dynamic"
	"github.com/hashicorp/packer-plugin-tester/datasource/parrot"
	"github.com/hashicorp/packer-plugin-tester/datasource/sleeper"
	dynamicPP "github.com/hashicorp/packer-plugin-tester/post-processor/dynamic"
	dynamicProv "github.com/hashicorp/packer-plugin-tester/provisioner/dynamic"
	"github.com/hashicorp/packer-plugin-tester/version"
)

func main() {
	pps := plugin.NewSet()
	pps.RegisterBuilder("dynamic", new(dynamic.Builder))
	pps.RegisterProvisioner("dynamic", new(dynamicProv.Provisioner))
	pps.RegisterPostProcessor("dynamic", new(dynamicPP.PostProcessor))
	pps.RegisterDatasource("dynamic", new(dynamicDS.Datasource))
	pps.RegisterDatasource("parrot", new(parrot.Datasource))
	pps.RegisterDatasource("sleeper", new(sleeper.Datasource))
	pps.SetVersion(version.PluginVersion)
	err := pps.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

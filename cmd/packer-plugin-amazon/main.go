package main

import (
	"fmt"
	"os"

	"github.com/hashicorp/packer-plugin-sdk/plugin"
	"github.com/hashicorp/packer/builder/amazon/ebs"
	"github.com/hashicorp/packer/builder/amazon/ebssurrogate"
	"github.com/hashicorp/packer/builder/amazon/ebsvolume"
	"github.com/hashicorp/packer/builder/osc/chroot"
	amazonami "github.com/hashicorp/packer/datasource/amazon/ami"
	"github.com/hashicorp/packer/datasource/amazon/secretsmanager"
	amazonimport "github.com/hashicorp/packer/post-processor/amazon-import"
)

func main() {
	pps := plugin.NewSet()
	pps.RegisterBuilder("ebs", new(ebs.Builder))
	pps.RegisterBuilder("chroot", new(chroot.Builder))
	pps.RegisterBuilder("ebssurrogate", new(ebssurrogate.Builder))
	pps.RegisterBuilder("ebsvolume", new(ebsvolume.Builder))
	pps.RegisterPostProcessor("import", new(amazonimport.PostProcessor))
	pps.RegisterDatasource("ami", new(amazonami.Datasource))
	pps.RegisterDatasource("secretsmanager", new(secretsmanager.Datasource))
	err := pps.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

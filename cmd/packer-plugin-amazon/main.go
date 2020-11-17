package main

import (
	"github.com/hashicorp/packer/builder/amazon/ebs"
	"github.com/hashicorp/packer/builder/amazon/ebssurrogate"
	"github.com/hashicorp/packer/builder/amazon/ebsvolume"
	"github.com/hashicorp/packer/builder/osc/chroot"
	"github.com/hashicorp/packer/packer/plugin"
	amazonimport "github.com/hashicorp/packer/post-processor/amazon-import"
)

func main() {
	plugin := plugin.New()
	plugin.RegisterBuilder("ebs", new(ebs.Builder))
	plugin.RegisterBuilder("chroot", new(chroot.Builder))
	plugin.RegisterBuilder("ebssurrogate", new(ebssurrogate.Builder))
	plugin.RegisterBuilder("ebsvolume", new(ebsvolume.Builder))
	plugin.RegisterPostProcessor("import", new(amazonimport.PostProcessor))
	err := plugin.Run()
	if err != nil {
		panic(err)
	}
}

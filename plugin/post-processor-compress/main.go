package main

import (
	"github.com/mitchellh/packer/packer/plugin"
	"github.com/mitchellh/packer/post-processor/compress"
)

func main() {
	plugin.ServePostProcessor(new(compress.PostProcessor))
}

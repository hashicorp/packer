package main

import (
	"github.com/mitchellh/packer/post-processor/compress"
	"github.com/mitchellh/packer/packer/plugin"
)

func main() {
	plugin.ServePostProcessor(new(compress.PostProcessor))
}

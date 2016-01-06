package plugin

import (
	"github.com/mitchellh/packer/packer"
	"log"
)

type cmdPostProcessor struct {
	p      packer.PostProcessor
	client *Client
}

func (c *cmdPostProcessor) Configure(config ...interface{}) error {
	defer func() {
		r := recover()
		c.checkExit(r, nil)
	}()

	return c.p.Configure(config...)
}

func (c *cmdPostProcessor) PostProcess(ui packer.Ui, a packer.Artifact) (packer.Artifact, bool, error) {
	defer func() {
		r := recover()
		c.checkExit(r, nil)
	}()

	return c.p.PostProcess(ui, a)
}

func (c *cmdPostProcessor) checkExit(p interface{}, cb func()) {
	if c.client.Exited() && cb != nil {
		cb()
	} else if p != nil && !Killed {
		log.Panic(p)
	}
}

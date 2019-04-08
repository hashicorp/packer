package plugin

import (
	"context"
	"log"

	"github.com/hashicorp/packer/packer"
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

func (c *cmdPostProcessor) PostProcess(ctx context.Context, ui packer.Ui, a packer.Artifact) (packer.Artifact, bool, bool, error) {
	defer func() {
		r := recover()
		c.checkExit(r, nil)
	}()

	return c.p.PostProcess(ctx, ui, a)
}

func (c *cmdPostProcessor) checkExit(p interface{}, cb func()) {
	if c.client.Exited() && cb != nil {
		cb()
	} else if p != nil && !Killed {
		log.Panic(p)
	}
}

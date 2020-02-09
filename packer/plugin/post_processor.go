package plugin

import (
	"context"
	"log"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer/packer"
)

type cmdPostProcessor struct {
	p      packer.PostProcessor
	client *Client
}

func (b *cmdPostProcessor) ConfigSpec() hcldec.ObjectSpec {
	defer func() {
		r := recover()
		b.checkExit(r, nil)
	}()

	return b.p.ConfigSpec()
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

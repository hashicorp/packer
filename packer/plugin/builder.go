package plugin

import (
	"context"
	"log"

	"github.com/hashicorp/packer/packer"
)

type cmdBuilder struct {
	builder packer.Builder
	client  *Client
}

func (b *cmdBuilder) Prepare(config ...interface{}) ([]string, error) {
	defer func() {
		r := recover()
		b.checkExit(r, nil)
	}()

	return b.builder.Prepare(config...)
}

func (b *cmdBuilder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {
	defer func() {
		r := recover()
		b.checkExit(r, nil)
	}()

	return b.builder.Run(ctx, ui, hook)
}

func (c *cmdBuilder) checkExit(p interface{}, cb func()) {
	if c.client.Exited() && cb != nil {
		cb()
	} else if p != nil && !Killed {
		log.Panic(p)
	}
}

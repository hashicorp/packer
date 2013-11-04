package plugin

import (
	"github.com/mitchellh/packer/packer"
	"log"
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

func (b *cmdBuilder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
	defer func() {
		r := recover()
		b.checkExit(r, nil)
	}()

	return b.builder.Run(ui, hook, cache)
}

func (b *cmdBuilder) Cancel() {
	defer func() {
		r := recover()
		b.checkExit(r, nil)
	}()

	b.builder.Cancel()
}

func (c *cmdBuilder) checkExit(p interface{}, cb func()) {
	if c.client.Exited() && cb != nil {
		cb()
	} else if p != nil && !Killed {
		log.Panic(p)
	}
}

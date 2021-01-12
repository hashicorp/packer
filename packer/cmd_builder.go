package packer

import (
	"context"
	"log"

	"github.com/hashicorp/hcl/v2/hcldec"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

type cmdBuilder struct {
	builder packersdk.Builder
	client  *PluginClient
}

func (b *cmdBuilder) ConfigSpec() hcldec.ObjectSpec {
	defer func() {
		r := recover()
		b.checkExit(r, nil)
	}()

	return b.builder.ConfigSpec()
}

func (b *cmdBuilder) Prepare(config ...interface{}) ([]string, []string, error) {
	defer func() {
		r := recover()
		b.checkExit(r, nil)
	}()

	return b.builder.Prepare(config...)
}

func (b *cmdBuilder) Run(ctx context.Context, ui packersdk.Ui, hook packersdk.Hook) (packersdk.Artifact, error) {
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

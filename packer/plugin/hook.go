package plugin

import (
	"context"
	"log"

	"github.com/hashicorp/packer/packer"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type cmdHook struct {
	hook   packer.Hook
	client *Client
}

func (c *cmdHook) Run(ctx context.Context, name string, ui packersdk.Ui, comm packer.Communicator, data interface{}) error {
	defer func() {
		r := recover()
		c.checkExit(r, nil)
	}()

	return c.hook.Run(ctx, name, ui, comm, data)
}

func (c *cmdHook) checkExit(p interface{}, cb func()) {
	if c.client.Exited() && cb != nil {
		cb()
	} else if p != nil && !Killed {
		log.Panic(p)
	}
}

package plugin

import (
	"context"
	"log"

	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type cmdHook struct {
	hook   packersdk.Hook
	client *Client
}

func (c *cmdHook) Run(ctx context.Context, name string, ui packersdk.Ui, comm packersdk.Communicator, data interface{}) error {
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

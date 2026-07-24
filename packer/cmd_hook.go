// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: BUSL-1.1

package packer

import (
	"context"
	"log"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

type cmdHook struct {
	hook   packersdk.Hook
	client *PluginClient
}

func (c *cmdHook) Run(ctx context.Context, name string, ui packersdk.Ui, comm packersdk.Communicator, data any) error {
	defer func() {
		r := recover()
		c.checkExit(r, nil)
	}()

	return c.hook.Run(ctx, name, ui, comm, data)
}

func (c *cmdHook) checkExit(p any, cb func()) {
	if c.client.Exited() && cb != nil {
		cb()
	} else if p != nil && !Killed {
		log.Panic(p)
	}
}

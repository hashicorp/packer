package plugin

import (
	"github.com/mitchellh/packer/packer"
	"log"
)

type cmdHook struct {
	hook   packer.Hook
	client *Client
}

func (c *cmdHook) Run(name string, ui packer.Ui, comm packer.Communicator, data interface{}) error {
	defer func() {
		r := recover()
		c.checkExit(r, nil)
	}()

	return c.hook.Run(name, ui, comm, data)
}

func (c *cmdHook) checkExit(p interface{}, cb func()) {
	if c.client.Exited() {
		cb()
	} else if p != nil {
		log.Panic(p)
	}
}

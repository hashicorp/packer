package plugin

import (
	"github.com/mitchellh/packer/packer"
	"log"
)

type cmdProvisioner struct {
	p      packer.Provisioner
	client *Client
}

func (c *cmdProvisioner) Prepare(configs ...interface{}) error {
	defer func() {
		r := recover()
		c.checkExit(r, nil)
	}()

	return c.p.Prepare(configs...)
}

func (c *cmdProvisioner) Provision(ui packer.Ui, comm packer.Communicator) error {
	defer func() {
		r := recover()
		c.checkExit(r, nil)
	}()

	return c.p.Provision(ui, comm)
}

func (c *cmdProvisioner) Cancel() {
	defer func() {
		r := recover()
		c.checkExit(r, nil)
	}()

	c.p.Cancel()
}

func (c *cmdProvisioner) checkExit(p interface{}, cb func()) {
	if c.client.Exited() && cb != nil {
		cb()
	} else if p != nil && !Killed {
		log.Panic(p)
	}
}

package plugin

import (
	"context"
	"log"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer/packer"
)

type cmdProvisioner struct {
	p      packer.Provisioner
	client *Client
}

func (p *cmdProvisioner) ConfigSpec() hcldec.ObjectSpec {
	defer func() {
		r := recover()
		p.checkExit(r, nil)
	}()

	return p.p.ConfigSpec()
}

func (c *cmdProvisioner) Prepare(configs ...interface{}) error {
	defer func() {
		r := recover()
		c.checkExit(r, nil)
	}()

	return c.p.Prepare(configs...)
}

func (c *cmdProvisioner) Provision(ctx context.Context, ui packer.Ui, comm packer.Communicator) error {
	defer func() {
		r := recover()
		c.checkExit(r, nil)
	}()

	return c.p.Provision(ctx, ui, comm)
}

func (c *cmdProvisioner) checkExit(p interface{}, cb func()) {
	if c.client.Exited() && cb != nil {
		cb()
	} else if p != nil && !Killed {
		log.Panic(p)
	}
}

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package packer

import (
	"context"
	"log"

	"github.com/hashicorp/hcl/v2/hcldec"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

type cmdProvisioner struct {
	p      packersdk.Provisioner
	client *PluginClient
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

func (c *cmdProvisioner) Provision(ctx context.Context, ui packersdk.Ui, comm packersdk.Communicator, generatedData map[string]interface{}) error {
	defer func() {
		r := recover()
		c.checkExit(r, nil)
	}()

	return c.p.Provision(ctx, ui, comm, generatedData)
}

func (c *cmdProvisioner) checkExit(p interface{}, cb func()) {
	if c.client.Exited() && cb != nil {
		cb()
	} else if p != nil && !Killed {
		log.Panic(p)
	}
}

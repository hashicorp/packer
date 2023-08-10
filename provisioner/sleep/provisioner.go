// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

//go:generate packer-sdc mapstructure-to-hcl2 -type Provisioner

package sleep

import (
	"context"
	"time"

	"github.com/hashicorp/hcl/v2/hcldec"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
)

type Provisioner struct {
	Duration time.Duration
}

var _ packersdk.Provisioner = new(Provisioner)

func (p *Provisioner) ConfigSpec() hcldec.ObjectSpec { return p.FlatMapstructure().HCL2Spec() }

func (p *Provisioner) FlatConfig() interface{} { return p.FlatMapstructure() }

func (p *Provisioner) Prepare(raws ...interface{}) error {
	return config.Decode(&p, &config.DecodeOpts{}, raws...)
}

func (p *Provisioner) Provision(ctx context.Context, _ packersdk.Ui, _ packersdk.Communicator, _ map[string]interface{}) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(p.Duration):
		return nil
	}
}

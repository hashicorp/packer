// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc mapstructure-to-hcl2 -type Config,NestedFirst,NestedSecond

package dynamic

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

type NestedSecond struct {
	Name string `mapstructure:"name" required:"true"`
}

type NestedFirst struct {
	Name    string         `mapstructure:"name" required:"true"`
	Nesteds []NestedSecond `mapstructure:"extra" required:"false"`
}

type Config struct {
	Nesteds []NestedFirst `mapstructure:"extra" required:"false"`
	ctx     interpolate.Context
}

type Provisioner struct {
	config Config
}

func (p *Provisioner) ConfigSpec() hcldec.ObjectSpec {
	return p.config.FlatMapstructure().HCL2Spec()
}

func (p *Provisioner) Prepare(raws ...interface{}) error {
	err := config.Decode(&p.config, &config.DecodeOpts{
		PluginType:         "packer.provisioner.dynamic",
		Interpolate:        true,
		InterpolateContext: &p.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{},
		},
	}, raws...)
	if err != nil {
		return err
	}
	return nil
}

func (p *Provisioner) Provision(_ context.Context, ui packer.Ui, _ packer.Communicator, generatedData map[string]interface{}) error {
	ui.Say(fmt.Sprintf("Called dynamic provisioner"))
	for _, nst := range p.config.Nesteds {
		ui.Say(fmt.Sprintf("Provisioner: nested one %s", nst.Name))
		for _, sec := range nst.Nesteds {
			ui.Say(fmt.Sprintf("Provisioner: nested second %s.%s", nst.Name, sec.Name))
		}
	}
	return nil
}

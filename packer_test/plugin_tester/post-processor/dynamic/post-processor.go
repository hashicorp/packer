// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc mapstructure-to-hcl2 -type Config,NestedFirst,NestedSecond

package dynamic

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcl/v2/hcldec"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
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

type PostProcessor struct {
	config Config
}

func (p *PostProcessor) ConfigSpec() hcldec.ObjectSpec { return p.config.FlatMapstructure().HCL2Spec() }

func (p *PostProcessor) Configure(raws ...interface{}) error {
	err := config.Decode(&p.config, &config.DecodeOpts{
		PluginType:         "packer.post-processor.dynamic",
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

func (p *PostProcessor) PostProcess(ctx context.Context, ui packersdk.Ui, source packersdk.Artifact) (packersdk.Artifact, bool, bool, error) {
	ui.Say(fmt.Sprintf("Called dynamic post-processor"))
	for _, nst := range p.config.Nesteds {
		ui.Say(fmt.Sprintf("Post-processor: nested one %s", nst.Name))
		for _, sec := range nst.Nesteds {
			ui.Say(fmt.Sprintf("Post-processor: nested second %s.%s", nst.Name, sec.Name))
		}
	}
	return source, true, true, nil
}

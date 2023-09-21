// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

//go:generate packer-sdc mapstructure-to-hcl2 -type Config

package breakpoint

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/common"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	Note    string `mapstructure:"note"`
	Disable bool   `mapstructure:"disable"`

	ctx interpolate.Context
}

type Provisioner struct {
	config Config
}

func (p *Provisioner) ConfigSpec() hcldec.ObjectSpec { return p.config.FlatMapstructure().HCL2Spec() }

func (p *Provisioner) Prepare(raws ...interface{}) error {
	err := config.Decode(&p.config, &config.DecodeOpts{
		PluginType:         "breakpoint",
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

func (p *Provisioner) Provision(ctx context.Context, ui packersdk.Ui, comm packersdk.Communicator, _ map[string]interface{}) error {
	if p.config.Disable {
		if p.config.Note != "" {
			ui.Say(fmt.Sprintf(
				"Breakpoint provisioner with note \"%s\" disabled; continuing...",
				p.config.Note))
		} else {
			ui.Say("Breakpoint provisioner disabled; continuing...")
		}

		return nil
	}
	if p.config.Note != "" {
		ui.Say(fmt.Sprintf("Pausing at breakpoint provisioner with note \"%s\".", p.config.Note))
	} else {
		ui.Say("Pausing at breakpoint provisioner.")
	}

	message := "Press enter to continue."

	var g errgroup.Group
	result := make(chan string, 1)
	g.Go(func() error {
		line, err := ui.Ask(message)
		if err != nil {
			return fmt.Errorf("Error asking for input: %s", err)
		}

		result <- line
		return nil
	})

	if err := g.Wait(); err != nil {
		return err
	}
	return nil
}

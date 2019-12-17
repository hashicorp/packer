//go:generate mapstructure-to-hcl2 -type Config

package breakpoint

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
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

func (p *Provisioner) Provision(ctx context.Context, ui packer.Ui, comm packer.Communicator) error {
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

	message := fmt.Sprintf(
		"Press enter to continue.")

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

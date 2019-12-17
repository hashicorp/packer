//go:generate mapstructure-to-hcl2 -type Provisioner

package sleep

import (
	"context"
	"time"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
)

type Provisioner struct {
	Duration time.Duration
}

var _ packer.Provisioner = new(Provisioner)

func (p *Provisioner) ConfigSpec() hcldec.ObjectSpec { return p.FlatMapstructure().HCL2Spec() }

func (p *Provisioner) FlatConfig() interface{} { return p.FlatMapstructure() }

func (p *Provisioner) Prepare(raws ...interface{}) error {
	return config.Decode(&p, &config.DecodeOpts{}, raws...)
}

func (p *Provisioner) Provision(ctx context.Context, _ packer.Ui, _ packer.Communicator) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(p.Duration):
		return nil
	}
}

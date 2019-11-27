package shell

import (
	"context"

	"github.com/hashicorp/hcl/v2/hcldec"
	sl "github.com/hashicorp/packer/common/shell-local"
	"github.com/hashicorp/packer/packer"
)

type Provisioner struct {
	config sl.Config
}

func (p *Provisioner) ConfigSpec() hcldec.ObjectSpec { return p.config.FlatMapstructure().HCL2Spec() }

func (p *Provisioner) Prepare(raws ...interface{}) error {
	err := sl.Decode(&p.config, raws...)
	if err != nil {
		return err
	}

	err = sl.Validate(&p.config)
	if err != nil {
		return err
	}

	return nil
}

func (p *Provisioner) Provision(ctx context.Context, ui packer.Ui, _ packer.Communicator) error {
	_, retErr := sl.Run(ctx, ui, &p.config)

	return retErr
}

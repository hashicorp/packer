package sleep

import (
	"context"
	"time"

	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
)

type Provisioner struct {
	Duration time.Duration
}

var _ packer.Provisioner = new(Provisioner)

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

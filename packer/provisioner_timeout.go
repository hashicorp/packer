package packer

import (
	"context"
	"fmt"
	"time"
)

// TimeoutProvisioner is a Provisioner implementation that can timeout after a
// duration
type TimeoutProvisioner struct {
	Provisioner
	Timeout time.Duration
}

func (p *TimeoutProvisioner) Provision(ctx context.Context, ui Ui, comm Communicator) error {
	ctx, cancel := context.WithTimeout(ctx, p.Timeout)
	defer cancel()

	// Use a select to determine if we get cancelled during the wait
	ui.Say(fmt.Sprintf("Setting a %s timeout for the next provisioner...", p.Timeout))
	return p.Provisioner.Provision(ctx, ui, comm)
}

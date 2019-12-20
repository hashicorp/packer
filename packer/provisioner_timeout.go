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

func (p *TimeoutProvisioner) Provision(ctx context.Context, ui Ui, comm Communicator, generatedData map[string]interface{}) error {
	ctx, cancel := context.WithTimeout(ctx, p.Timeout)
	defer cancel()

	// Use a select to determine if we get cancelled during the wait
	ui.Say(fmt.Sprintf("Setting a %s timeout for the next provisioner...", p.Timeout))

	errC := make(chan interface{})

	go func() {
		select {
		case <-errC:
			// all good
		case <-ctx.Done():
			switch ctx.Err() {
			case context.DeadlineExceeded:
				ui.Error("Cancelling provisioner after a timeout...")
			default:
				// the context also gets cancelled when the provisioner is
				// successful
			}
		}
	}()

	err := p.Provisioner.Provision(ctx, ui, comm, generatedData)
	close(errC)
	return err
}

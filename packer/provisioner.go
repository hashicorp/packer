package packer

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/hashicorp/hcl/v2/hcldec"
)

// A provisioner is responsible for installing and configuring software
// on a machine prior to building the actual image.
type Provisioner interface {
	HCL2Speccer

	// Prepare is called with a set of configurations to setup the
	// internal state of the provisioner. The multiple configurations
	// should be merged in some sane way.
	Prepare(...interface{}) error

	// Provision is called to actually provision the machine. A context is
	// given for cancellation, a UI is given to communicate with the user, and
	// a communicator is given that is guaranteed to be connected to some
	// machine so that provisioning can be done.
	Provision(context.Context, Ui, Communicator) error
}

// A HookedProvisioner represents a provisioner and information describing it
type HookedProvisioner struct {
	Provisioner Provisioner
	Config      interface{}
	TypeName    string
}

// A Hook implementation that runs the given provisioners.
type ProvisionHook struct {
	// The provisioners to run as part of the hook. These should already
	// be prepared (by calling Prepare) at some earlier stage.
	Provisioners []*HookedProvisioner
}

// Runs the provisioners in order.
func (h *ProvisionHook) Run(ctx context.Context, name string, ui Ui, comm Communicator, data interface{}) error {
	// Shortcut
	if len(h.Provisioners) == 0 {
		return nil
	}

	if comm == nil {
		return fmt.Errorf(
			"No communicator found for provisioners! This is usually because the\n" +
				"`communicator` config was set to \"none\". If you have any provisioners\n" +
				"then a communicator is required. Please fix this to continue.")
	}
	for _, p := range h.Provisioners {
		ts := CheckpointReporter.AddSpan(p.TypeName, "provisioner", p.Config)

		err := p.Provisioner.Provision(ctx, ui, comm)

		ts.End(err)
		if err != nil {
			return err
		}
	}

	return nil
}

// PausedProvisioner is a Provisioner implementation that pauses before
// the provisioner is actually run.
type PausedProvisioner struct {
	PauseBefore time.Duration
	Provisioner Provisioner
}

func (p *PausedProvisioner) ConfigSpec() hcldec.ObjectSpec { return p.ConfigSpec() }
func (p *PausedProvisioner) FlatConfig() interface{}       { return p.FlatConfig() }
func (p *PausedProvisioner) Prepare(raws ...interface{}) error {
	return p.Provisioner.Prepare(raws...)
}

func (p *PausedProvisioner) Provision(ctx context.Context, ui Ui, comm Communicator) error {

	// Use a select to determine if we get cancelled during the wait
	ui.Say(fmt.Sprintf("Pausing %s before the next provisioner...", p.PauseBefore))
	select {
	case <-time.After(p.PauseBefore):
	case <-ctx.Done():
		return ctx.Err()
	}

	return p.Provisioner.Provision(ctx, ui, comm)
}

// DebuggedProvisioner is a Provisioner implementation that waits until a key
// press before the provisioner is actually run.
type DebuggedProvisioner struct {
	Provisioner Provisioner

	cancelCh chan struct{}
	doneCh   chan struct{}
	lock     sync.Mutex
}

func (p *DebuggedProvisioner) ConfigSpec() hcldec.ObjectSpec { return p.ConfigSpec() }
func (p *DebuggedProvisioner) FlatConfig() interface{}       { return p.FlatConfig() }
func (p *DebuggedProvisioner) Prepare(raws ...interface{}) error {
	return p.Provisioner.Prepare(raws...)
}

func (p *DebuggedProvisioner) Provision(ctx context.Context, ui Ui, comm Communicator) error {
	// Use a select to determine if we get cancelled during the wait
	message := "Pausing before the next provisioner . Press enter to continue."

	result := make(chan string, 1)
	go func() {
		line, err := ui.Ask(message)
		if err != nil {
			log.Printf("Error asking for input: %s", err)
		}

		result <- line
	}()

	select {
	case <-result:
	case <-ctx.Done():
		return ctx.Err()
	}

	return p.Provisioner.Provision(ctx, ui, comm)
}

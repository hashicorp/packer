package packer

import (
	"fmt"
	"sync"
	"time"
)

// A provisioner is responsible for installing and configuring software
// on a machine prior to building the actual image.
type Provisioner interface {
	// Prepare is called with a set of configurations to setup the
	// internal state of the provisioner. The multiple configurations
	// should be merged in some sane way.
	Prepare(...interface{}) error

	// Provision is called to actually provision the machine. A UI is
	// given to communicate with the user, and a communicator is given that
	// is guaranteed to be connected to some machine so that provisioning
	// can be done.
	Provision(Ui, Communicator) error

	// Cancel is called to cancel the provisioning. This is usually called
	// while Provision is still being called. The Provisioner should act
	// to stop its execution as quickly as possible in a race-free way.
	Cancel()
}

// A Hook implementation that runs the given provisioners.
type ProvisionHook struct {
	// The provisioners to run as part of the hook. These should already
	// be prepared (by calling Prepare) at some earlier stage.
	Provisioners []Provisioner

	lock               sync.Mutex
	runningProvisioner Provisioner
}

// Runs the provisioners in order.
func (h *ProvisionHook) Run(name string, ui Ui, comm Communicator, data interface{}) error {
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

	defer func() {
		h.lock.Lock()
		defer h.lock.Unlock()

		h.runningProvisioner = nil
	}()

	for _, p := range h.Provisioners {
		h.lock.Lock()
		h.runningProvisioner = p
		h.lock.Unlock()

		if err := p.Provision(ui, comm); err != nil {
			return err
		}
	}

	return nil
}

// Cancels the privisioners that are still running.
func (h *ProvisionHook) Cancel() {
	h.lock.Lock()
	defer h.lock.Unlock()

	if h.runningProvisioner != nil {
		h.runningProvisioner.Cancel()
	}
}

// PausedProvisioner is a Provisioner implementation that pauses before
// the provisioner is actually run.
type PausedProvisioner struct {
	PauseBefore time.Duration
	Provisioner Provisioner

	cancelCh chan struct{}
	doneCh   chan struct{}
	lock     sync.Mutex
}

func (p *PausedProvisioner) Prepare(raws ...interface{}) error {
	return p.Provisioner.Prepare(raws...)
}

func (p *PausedProvisioner) Provision(ui Ui, comm Communicator) error {
	p.lock.Lock()
	cancelCh := make(chan struct{})
	p.cancelCh = cancelCh

	// Setup the done channel, which is trigger when we're done
	doneCh := make(chan struct{})
	defer close(doneCh)
	p.doneCh = doneCh
	p.lock.Unlock()

	defer func() {
		p.lock.Lock()
		defer p.lock.Unlock()
		if p.cancelCh == cancelCh {
			p.cancelCh = nil
		}
		if p.doneCh == doneCh {
			p.doneCh = nil
		}
	}()

	// Use a select to determine if we get cancelled during the wait
	ui.Say(fmt.Sprintf("Pausing %s before the next provisioner...", p.PauseBefore))
	select {
	case <-time.After(p.PauseBefore):
	case <-cancelCh:
		return nil
	}

	provDoneCh := make(chan error, 1)
	go p.provision(provDoneCh, ui, comm)

	select {
	case err := <-provDoneCh:
		return err
	case <-cancelCh:
		p.Provisioner.Cancel()
		return <-provDoneCh
	}
}

func (p *PausedProvisioner) Cancel() {
	var doneCh chan struct{}

	p.lock.Lock()
	if p.cancelCh != nil {
		close(p.cancelCh)
		p.cancelCh = nil
	}
	if p.doneCh != nil {
		doneCh = p.doneCh
	}
	p.lock.Unlock()

	<-doneCh
}

func (p *PausedProvisioner) provision(result chan<- error, ui Ui, comm Communicator) {
	result <- p.Provisioner.Provision(ui, comm)
}

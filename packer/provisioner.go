package packer

import (
	"sync"
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

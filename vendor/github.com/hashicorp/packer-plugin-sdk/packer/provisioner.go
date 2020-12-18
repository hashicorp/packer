package packer

import "context"

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
	Provision(context.Context, Ui, Communicator, map[string]interface{}) error
}

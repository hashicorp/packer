package packer

// A provisioner is responsible for installing and configuring software
// on a machine prior to building the actual image.
type Provisioner interface {
	// Prepare is called with the raw configuration and a UI element in
	// order to setup the internal state of the provisioner and perform
	// any validation necessary for the provisioner.
	Prepare(interface{}, Ui)

	// Provision is called to actually provision the machine. A UI is
	// given to communicate with the user, and a communicator is given that
	// is guaranteed to be connected to some machine so that provisioning
	// can be done.
	Provision(Ui, Communicator)
}

package packer

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
}

// A Hook implementation that runs the given provisioners.
type ProvisionHook struct {
	// The provisioners to run as part of the hook. These should already
	// be prepared (by calling Prepare) at some earlier stage.
	Provisioners []Provisioner
}

// Runs the provisioners in order.
func (h *ProvisionHook) Run(name string, ui Ui, comm Communicator, data interface{}) error {
	for _, p := range h.Provisioners {
		if err := p.Provision(ui, comm); err != nil {
			return err
		}
	}

	return nil
}

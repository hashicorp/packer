package classic

import (
	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
)

type stepCreateIPReservation struct{}

func (s *stepCreateIPReservation) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	ui.Say("Creating IP reservation...")
	const endpoint_path = "/ip/reservation/" // POST

	// $ opc compute ip-reservations add \
	//     /Compute-mydomain/user@example.com/master-instance-ip \
	//     /oracle/public/ippool

	// account /Compute-mydomain/default
	// ip      129.144.27.172
	// name    /Compute-mydomain/user@example.com/master-instance-ip
	// ...

}

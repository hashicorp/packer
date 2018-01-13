package classic

import (
	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
)

type stepCreateIPReservation struct{}

func (s *stepInstanceInfo) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	ui.Say("Getting Instance Info...")
	endpoint_path := "/instance/%s", instanceName // GET

	// $ opc compute ip-reservations add \
	//     /Compute-mydomain/user@example.com/master-instance-ip \
	//     /oracle/public/ippool

	// account /Compute-mydomain/default
	// ip      129.144.27.172
	// name    /Compute-mydomain/user@example.com/master-instance-ip
	// ...

}

package classic

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-oracle-terraform/compute"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepCreateIPReservation struct{}

func (s *stepCreateIPReservation) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	config := state.Get("config").(*Config)
	client := state.Get("client").(*compute.ComputeClient)
	iprClient := client.IPReservations()
	// TODO: add optional Name and Tags

	ipresName := fmt.Sprintf("ipres_%s", config.ImageName)
	ui.Say(fmt.Sprintf("Creating IP reservation: %s", ipresName))

	IPInput := &compute.CreateIPReservationInput{
		ParentPool: compute.PublicReservationPool,
		Permanent:  true,
		Name:       ipresName,
	}
	ipRes, err := iprClient.CreateIPReservation(IPInput)

	if err != nil {
		err := fmt.Errorf("Error creating IP Reservation: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	state.Put("instance_ip", ipRes.IP)
	state.Put("ipres_name", ipresName)
	return multistep.ActionContinue
}

func (s *stepCreateIPReservation) Cleanup(state multistep.StateBag) {
	// TODO: delete ip reservation
}

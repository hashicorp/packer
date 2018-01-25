package classic

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/go-oracle-terraform/compute"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepCreateIPReservation struct{}

func (s *stepCreateIPReservation) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	ui.Say("Creating IP reservation...")
	config := state.Get("config").(*Config)
	client := state.Get("client").(*compute.ComputeClient)
	iprClient := client.IPReservations()
	// TODO: add optional Name and Tags
	IPInput := &compute.CreateIPReservationInput{
		ParentPool: compute.PublicReservationPool,
		Permanent:  true,
		Name:       fmt.Sprintf("ipres_%s", config.ImageName),
	}
	ipRes, err := iprClient.CreateIPReservation(IPInput)

	if err != nil {
		err := fmt.Errorf("Error creating IP Reservation: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	state.Put("instance_ip", ipRes.IP)
	log.Printf("debug: ipRes is %#v", ipRes)
	return multistep.ActionContinue
}

func (s *stepCreateIPReservation) Cleanup(state multistep.StateBag) {
	// TODO: delete ip reservation
}

package classic

import (
	"log"

	"github.com/hashicorp/go-oracle-terraform/compute"
	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
)

type stepCreateIPReservation struct{}

func (s *stepCreateIPReservation) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	ui.Say("Creating IP reservation...")
	client := state.Get("client", client).(*compute.ComputeClient)
	iprClient := client.IPReservations()
	if err != nil {
		log.Printf("Error getting IPReservations Client: %s", err)
		return multistep.ActionHalt
	}
	// TODO: add optional Name and Tags
	IPInput := &iprClient.CreateIPReservationInput{
		ParentPool: compute.PublicReservationPool,
		Permanent:  true,
	}
	ipRes, err := iprClient.CreateIPReservation(createIPReservation)
	if err != nil {
		log.Printf("Error creating IP Reservation: %s", err)
		return multistep.ActionHalt
	}
	log.Printf("debug: ipRes is %#v", ipRes)
	return multistep.ActionContinue
}
